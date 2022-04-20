package lan

import (
	"fmt"
	"strings"

	"github.com/MarcPer/lanchat/ui"
)

type InboundHandler func(*Client, Packet, peerID)
type OutboundHandler func(*Client, ui.Packet)
type MsgHandler struct {
	in  InboundHandler
	out OutboundHandler
}

var MsgHandlers = map[string]MsgHandler{
	":id": {idInHandler, idOutHandler},
}

func idInHandler(c *Client, p Packet, from peerID) {
	args := strings.Split(p.Msg, " ")
	if len(args) != 2 || args[1] == "" {
		c.ToUI <- ui.Packet{Type: ui.PacketTypeAdmin, Msg: fmt.Sprintf(":id needs a single, non-empty argument, received %v\n", args[1:])}
		return
	}
	peersMu.RLock()
	defer peersMu.RUnlock()
	if peer, ok := c.peers[from]; ok {
		var msg string
		if peer.name == "" {
			msg = fmt.Sprintf("user \"%s\" connected", args[1])
			c.transmit(Packet{Type: MsgTypeCmd, Msg: ":id " + c.Name}, from)
		} else if peer.name == args[1] {
			// nothing to do
			return
		} else {
			msg = fmt.Sprintf("user \"%s\" changed their name to \"%s\"", peer.name, args[1])
		}
		peer.name = args[1]
		c.ToUI <- ui.Packet{Msg: msg, Type: ui.PacketTypeAdmin}
		c.broadcast(Packet{Type: MsgTypeAdmin, Msg: msg}, from)
	}

}

func idOutHandler(c *Client, p ui.Packet) {
	args := strings.Split(p.Msg, " ")
	if len(args) != 2 || args[1] == "" {
		c.logToUIf(":id needs a single, non-empty argument, received %v\n", args[1:])
		return
	}
	c.broadcast(Packet{User: c.Name, Msg: p.Msg, Type: MsgTypeCmd}, "")
	c.Name = args[1]
	go func() {
		c.ToUI <- ui.Packet{Type: ui.PacketTypeCmd, Msg: p.Msg}
	}()
}

func handleInbound(c *Client, p Packet, from peerID) {
	switch p.Type {
	case MsgTypeChat:
		c.ToUI <- ui.Packet{User: p.User, Msg: p.Msg}
		c.broadcast(p, from)
		return
	case MsgTypeAdmin:
		c.ToUI <- ui.Packet{User: p.User, Msg: p.Msg, Type: ui.PacketTypeAdmin}
	case MsgTypeCmd:
		if h, ok := checkInCmd(p.Msg); ok {
			h(c, p, from)
		} else {
			c.logToUIf("invalid command '%s'. Run ':h' or ':help' to see available commands\n", p.Msg)
		}
		return
	}
}

func checkInCmd(msg string) (InboundHandler, bool) {
	if !strings.HasPrefix(msg, ":") {
		return nil, false
	}
	args := strings.Split(msg, " ")
	h, ok := MsgHandlers[args[0]]
	return h.in, ok
}

func handleOutbound(c *Client, p ui.Packet) {
	if strings.HasPrefix(p.Msg, ":") {
		if h, ok := checkOutCmd(p.Msg); ok {
			h(c, p)
		} else {
			c.logToUIf("invalid command '%s'. Run ':h' or ':help' to see available commands\n", p.Msg)
		}
	} else {
		c.broadcast(Packet{User: c.Name, Msg: p.Msg, Type: MsgTypeChat}, "")
	}
}

func checkOutCmd(msg string) (OutboundHandler, bool) {
	args := strings.Split(msg, " ")
	h, ok := MsgHandlers[args[0]]
	return h.out, ok
}
