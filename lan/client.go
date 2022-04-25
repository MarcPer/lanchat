package lan

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
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
)

type Packet struct {
	User string
	Msg  string
	Type int
}

type peer struct {
	name string
	conn *net.Conn
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
}

func (c *Client) Start(ctx context.Context) {
	c.peers = make(map[peerID]*peer)
	host, found := c.Scanner.FindHost(c.HostPort)
	c.host = !found
	if found { // host found, so become regular peer
		logger.Infof("Found host at %s; connecting... \n", host)
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logger.Errorf("Could not connect to host: %v\n", err)
			os.Exit(1)
		}
		var pid peerID = peerID(conn.RemoteAddr().String())
		enc := gob.NewEncoder(conn)
		peersMu.Lock()
		defer peersMu.Unlock()
		c.peers[pid] = &peer{conn: &conn, enc: enc}
		c.transmit(Packet{User: "", Type: MsgTypeCmd, Msg: ":id " + c.Name}, pid)
		go c.handleConn(pid)
	} else { // become a host
		logger.Infof("No host found; starting server in 0.0.0.0:%d ... \n", c.HostPort)
		time.Sleep(2 * time.Second)
		go c.serve(ctx)
	}

	go c.handleUIPackets(ctx)
}

func (c *Client) serve(ctx context.Context) {
	url := fmt.Sprintf(":%d", c.HostPort)
	ln, err := net.Listen("tcp", url)
	if err != nil {
		logger.Errorf("Could not start server: %v\n", err)
		os.Exit(1)
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
			c.peers[pid] = &peer{conn: &conn, enc: enc}
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
	dec := gob.NewDecoder(*peer.conn)
	for {
		var pkt Packet
		err := dec.Decode(&pkt)
		if err == io.EOF {
			if peer.name != "" {
				logger.Infof("connection closed by peer %s\n", pid)
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
			close(c.ToUI)
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

func (c *Client) processCommand(pkt Packet, pid peerID) {
	if !strings.HasPrefix(pkt.Msg, ":") {
		logger.Warnf("invalid command: %v\n", pkt.Msg)
		return
	}
	args := strings.Split(pkt.Msg, " ")

	switch args[0] {
	case ":id":
		if len(args) != 2 || args[1] == "" {
			logger.Warnf(":id needs a single, non-empty argument, received %v\n", args[1:])
			return
		}
		peersMu.RLock()
		defer peersMu.RUnlock()
		if peer, ok := c.peers[pid]; ok {
			var msg string
			if peer.name == "" {
				msg = fmt.Sprintf("user \"%s\" connected", args[1])
				c.transmit(Packet{Type: MsgTypeCmd, Msg: ":id " + c.Name}, pid)
			} else if peer.name == args[1] {
				// nothing to do
				return
			} else {
				msg = fmt.Sprintf("user \"%s\" changed their name to \"%s\"", peer.name, args[1])
			}
			peer.name = args[1]
			c.ToUI <- ui.Packet{Msg: msg, Type: ui.PacketTypeAdmin}
		}
	}
}

func (c *Client) cleanPeer(pid peerID) {
	peersMu.Lock()
	delete(c.peers, pid)
	peersMu.Unlock()
}

func (c *Client) logToUIf(format string, v ...interface{}) {
	c.ToUI <- ui.Packet{Type: ui.PacketTypeAdmin, Msg: fmt.Sprintf(format, v...)}
}
