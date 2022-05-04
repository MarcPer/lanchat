package lan

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/MarcPer/lanchat/logger"
	"github.com/MarcPer/lanchat/ui"
)

type peerID string

var peersMu sync.RWMutex

type MsgType int

const (
	MsgTypeChat = iota
	MsgTypeCmd
	MsgTypeAdmin
	MsgTypePing
)

type Packet struct {
	User string
	Msg  string
	Type int
}

type peer struct {
	name string
	conn io.Reader
	enc  *gob.Encoder
}

type Client struct {
	Name     string
	HostPort int
	ToUI     chan ui.Packet
	FromUI   chan ui.Packet
	Scanner  NetScanner
	host     bool
	peers    map[peerID]*peer
	ctx      context.Context
	cancel   context.CancelFunc
	restart  chan int
}

func (c *Client) Start(ctx context.Context) {
	if ctx != nil {
		c.ctx = ctx
	}
	c.restart = make(chan int)
	go c.monitor()
	c.retry(0)
}

func (c *Client) run(ctx context.Context) {
	c.peers = make(map[peerID]*peer)
	c.logToUIf("Scanning for hosts")
	host, found := c.Scanner.FindHost(c.HostPort)
	c.host = !found
	if found { // host found, so become regular peer
		c.logToUIf("Found host at %s; connecting...", host)
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logger.Errorf("Could not connect to host: %v\n", err)
			c.retry(-1)
			return
		}
		var pid peerID = peerID(conn.RemoteAddr().String())
		enc := gob.NewEncoder(conn)
		peersMu.Lock()
		c.peers[pid] = &peer{conn: conn, enc: enc}
		peersMu.Unlock()
		c.transmit(Packet{User: "", Type: MsgTypeCmd, Msg: ":id " + c.Name}, pid)
		go c.handleConn(pid)
	} else { // become a host
		c.logToUIf("No host found; starting server at 0.0.0.0:%d ...", c.HostPort)
		go c.serve(ctx)
	}

	go c.handleUIPackets(ctx)
	go c.ping(ctx)
}

func (c *Client) monitor() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case t := <-c.restart:
			if c.cancel != nil {
				c.cancel()
			}
			subCtx, cancel := context.WithCancel(c.ctx)
			c.cancel = cancel
			c.host = true // this prevents disconnections from triggering another, possibly concurrent, restart
			// check the cleanPeer method to understand why.

			// Wait a random time before starting again, to avoid two peers
			// trying to become a host simultaneously
			// A more reliable scheme should be used, in which peers
			// know about the existence of others, not just the host.
			// They can then coordinate better in such cases.
			time.Sleep(time.Duration(t) * time.Millisecond)
			go c.run(subCtx)
		}
	}
}

// takes time to wait before restarting, in milliseconds
// if < 0, sets it to be a random value
func (c *Client) retry(t int) {
	if t < 0 {
		t = rand.Intn(8000)
	}
	c.restart <- t
}

func (c *Client) serve(ctx context.Context) {
	url := fmt.Sprintf(":%d", c.HostPort)
	ln, err := net.Listen("tcp", url)
	if err != nil {
		logger.Errorf("Could not start server: %v\n", err)
		c.retry(-1)
		return
	}

	connCh := make(chan net.Conn)

	go func(l net.Listener, ch chan net.Conn) {
		for {
			conn, err := l.Accept()
			if err != nil {
				logger.Debugf("serve: %v\n", err)
				return
			} else {
				logger.Debugf("new connection from %v\n", conn.RemoteAddr().String())
				ch <- conn
			}
		}
	}(ln, connCh)

	for {
		select {
		case conn := <-connCh:
			var pid peerID = peerID(conn.RemoteAddr().String())
			peersMu.Lock()
			enc := gob.NewEncoder(conn)
			c.peers[pid] = &peer{conn: conn, enc: enc}
			peersMu.Unlock()
			go c.handleConn(pid)
		case <-ctx.Done():
			ln.Close()
			return
		}
	}
}

func (c *Client) handleConn(pid peerID) {
	peersMu.Lock()
	peer, ok := c.peers[pid]
	peersMu.Unlock()
	if !ok {
		logger.Debugf("handleConn: peer with ID=%v not found\n", pid)
		return
	}
	dec := gob.NewDecoder(peer.conn)
	for {
		var pkt Packet
		err := dec.Decode(&pkt)
		if err == io.EOF {
			if peer.name != "" {
				msg := fmt.Sprintf("'%s' disconnected\n", peer.name)
				c.logToUI(msg)
				c.broadcast(Packet{Msg: msg, Type: MsgTypeAdmin}, pid)
			}
			c.cleanPeer(pid)
			return
		} else if err != nil {
			logger.Errorf("handleConn: error decoding packet %v\n", err)
			// return
		} else {
			handleInbound(c, pkt, pid)
		}
	}
}

func (c *Client) handleUIPackets(ctx context.Context) {
	for {
		select {
		case p := <-c.FromUI:
			logger.Debugf("p=%+v, msg=%q\n", p, p.Msg)
			handleOutbound(c, p)
		case <-ctx.Done():
			return
		}
	}
}

var pingPacket = Packet{Type: MsgTypePing}

func (c *Client) ping(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.broadcast(pingPacket, "")
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) broadcast(pkt Packet, except peerID) {
	peersMu.RLock()
	for pid := range c.peers {
		if pid == except {
			continue
		}
		c.transmit(pkt, pid)
	}
	peersMu.RUnlock()
}

func (c *Client) transmit(pkt Packet, pid peerID) {
	peer, ok := c.peers[pid]
	if !ok {
		logger.Warnf("transmit: peer with ID=%v not found\n", pid)
		return
	}
	if err := peer.enc.Encode(pkt); err != nil {
		logger.Errorf("transmit: error encoding packet %v\n", err)
		// failed to send data to peer
		c.cleanPeer(pid)
	}
}

func (c *Client) cleanPeer(pid peerID) {
	peersMu.Lock()
	defer peersMu.Unlock()
	delete(c.peers, pid)

	if !c.host && len(c.peers) < 1 {
		c.retry(-1)
	}
}

func (c *Client) logToUI(msg string) {
	c.ToUI <- ui.Packet{Type: ui.PacketTypeAdmin, Msg: msg}
}

func (c *Client) logToUIf(format string, v ...interface{}) {
	c.ToUI <- ui.Packet{Type: ui.PacketTypeAdmin, Msg: fmt.Sprintf(format, v...)}
}
