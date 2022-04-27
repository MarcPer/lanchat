package lan

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/MarcPer/lanchat/ui"
)

func newTestClient(host bool, numPeers int, scanner NetScanner) Client {
	peers := make(map[peerID]*peer)

	for i := 0; i < numPeers; i++ {
		id := strconv.Itoa(i)
		var buf bytes.Buffer
		p := peer{
			name: "peer_" + id,
			conn: &buf,
			enc:  gob.NewEncoder(&buf),
		}
		peers[peerID(id)] = &p
	}
	return Client{
		Name:     "testClient",
		HostPort: 6776,
		ToUI:     make(chan ui.Packet, 10),
		FromUI:   make(chan ui.Packet, 10),
		peers:    peers,
	}
}

func readFromPeer(c *Client, peerIdx int) (out []Packet, err error) {
	out = make([]Packet, 0)
	pid := peerID(strconv.Itoa(peerIdx))
	peer, ok := c.peers[pid]
	if !ok {
		err = fmt.Errorf("no peer found with id=%d", peerIdx)
		return
	}
	dec := gob.NewDecoder(peer.conn)
	for {
		var pkt Packet
		e := dec.Decode(&pkt)
		if e == io.EOF {
			return
		} else if err != nil {
			e = fmt.Errorf("error decoding peer packet: %v", err)
			return
		} else {
			out = append(out, pkt)
		}
	}
}

// reads all UI packets received. Note: This function closes the UI channel
func readUI(c *Client) (out []ui.Packet, err error) {
	out = make([]ui.Packet, 0)
	close(c.ToUI)
	for {
		select {
		case p, ok := <-c.ToUI:
			if !ok {
				return
			}
			out = append(out, p)
		case <-time.After(time.Second):
			return
		}
	}
}
