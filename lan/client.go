package lan

import (
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/MarcPer/lanchat/ui"
)

type peerID string

var peersMu sync.Mutex

type MsgType int

const (
	MsgTypeChat = iota
	MsgTypeCmd
)

type Packet struct {
	Msg  string
	Type MsgType
}

type peer struct {
	name string
	conn net.Conn
}

type Client struct {
	Name     string
	Local    bool
	HostPort int
	ToUI     chan ui.Packet
	FromUI   chan ui.Packet
	Scanner  NetScanner
	host     bool
	peers    map[peerID]peer
}

func (c *Client) Start(ctx context.Context) {
	c.peers = make(map[peerID]peer)
	host, found := c.Scanner.FindHost(c.HostPort)
	c.host = !found
	if found { // host found, so become regular peer
		url := fmt.Sprintf("%s:%d", host, c.HostPort)
		conn, err := net.Dial("tcp", url)
		if err != nil {
			fmt.Printf("Could not connect to host: %v\n", err)
			os.Exit(1)
		}
		var pid peerID = peerID(conn.RemoteAddr().String())
		c.peers[pid] = peer{name: conn.RemoteAddr().String(), conn: conn}
		go c.handleConn(pid)
	} else { // become a host
		go c.serve(ctx)
	}

	c.handleUIPackets(ctx)
}

func (c *Client) serve(ctx context.Context) {
	url := fmt.Sprintf(":%d", c.HostPort)
	ln, err := net.Listen("tcp", url)
	if err != nil {
		fmt.Printf("Could not start server: %v\n", err)
		os.Exit(1)
	}

	connCh := make(chan net.Conn)

	go func(l net.Listener, ch chan net.Conn) {
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Printf("Failed to accept connection: %v", err)
				return
			} else {
				ch <- conn
			}
		}
	}(ln, connCh)

	for {
		select {
		case conn := <-connCh:
			var pid peerID = peerID(conn.RemoteAddr().String())
			peersMu.Lock()
			c.peers[pid] = peer{conn: conn}
			peersMu.Unlock()
			go c.handleConn(pid)
		case <-ctx.Done():
			ln.Close()
			return
		}
	}
}

func (c *Client) handleConn(pid peerID) {
	peer, ok := c.peers[pid]
	if !ok {
		fmt.Printf("handleConn: peer with ID=%v not found\n", pid)
		return
	}
	dec := gob.NewDecoder(peer.conn)
	for {
		var pkt Packet
		err := dec.Decode(&pkt)
		if err != nil {
			fmt.Printf("handleConn: error decoding packet %v\n", err)
			return
		} else {
			c.ToUI <- ui.Packet{User: peer.name, Msg: pkt.Msg}
		}
	}
}

func (c *Client) handleUIPackets(ctx context.Context) {
	for {
		select {
		case p := <-c.FromUI:
			c.broadcast(Packet{Msg: p.Msg})
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) broadcast(pkt Packet) {
	for pid := range c.peers {
		c.transmit(pkt, pid)
	}
}

func (c *Client) transmit(pkt Packet, pid peerID) {
	peer := c.peers[pid]
	enc := gob.NewEncoder(peer.conn)
	if err := enc.Encode(pkt); err != nil {
		// failed to send data to peer
		peersMu.Lock()
		delete(c.peers, pid)
		peersMu.Unlock()
	}
}
