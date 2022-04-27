package lan

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/MarcPer/lanchat/ui"
)

func TestCheckInCmd(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		ok   bool
	}{
		{"empty message", "", false},
		{"wrongly formatted command", "id", false},
		{":id command", ":id", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := checkInCmd(tt.msg)
			if ok != tt.ok {
				t.Errorf("expected ok=%v, got=%v", tt.ok, ok)
			}
		})

	}
}

func TestHandleInbound(t *testing.T) {
	tests := []struct {
		name        string
		from        peerID
		in          Packet
		uiPackets   []ui.Packet
		peerPackets [][]Packet
	}{
		{
			"chat message from peer 0",
			"0",
			Packet{User: "peer_0", Msg: "test"},
			[]ui.Packet{{User: "peer_0", Msg: "test", Type: ui.PacketTypeChat}},
			[][]Packet{
				{},
				{Packet{User: "peer_0", Msg: "test"}},
			},
		},
		{
			"chat message from peer 1",
			"1",
			Packet{User: "peer_1", Msg: "test"},
			[]ui.Packet{{User: "peer_1", Msg: "test", Type: ui.PacketTypeChat}},
			[][]Packet{
				{Packet{User: "peer_1", Msg: "test"}},
				{},
			},
		},
		{
			"admin message",
			"0",
			Packet{User: "peer_0", Msg: "connected", Type: MsgTypeAdmin},
			[]ui.Packet{{User: "peer_0", Msg: "connected", Type: ui.PacketTypeAdmin}},
			[][]Packet{
				{},
				{},
			},
		},
		{
			"invalid command",
			"0",
			Packet{User: "peer_0", Msg: ":fake_cmd", Type: MsgTypeCmd},
			[]ui.Packet{{User: "", Msg: "invalid command ':fake_cmd'. Run ':h' or ':help' to see available commands\n", Type: ui.PacketTypeAdmin}},
			[][]Packet{
				{},
				{},
			},
		},
		{
			":id command",
			"0",
			Packet{User: "peer_0", Msg: ":id jon", Type: MsgTypeCmd},
			[]ui.Packet{{User: "", Msg: "user \"peer_0\" changed their name to \"jon\"", Type: ui.PacketTypeAdmin}},
			[][]Packet{
				{},
				{Packet{Type: MsgTypeAdmin, Msg: "user \"peer_0\" changed their name to \"jon\""}},
			},
		},
	}

	numPeers := 2
	for _, tt := range tests {
		c := newTestClient(true, numPeers, &NullScanner{})
		t.Run(tt.name, func(t *testing.T) {
			// check if test is setup properly
			if len(tt.peerPackets) != numPeers {
				t.Fatalf("client has %d peers, but test setup has only %d", numPeers, len(tt.peerPackets))
			}

			handleInbound(&c, tt.in, tt.from)

			// check received UI packets
			uiPackets, err := readUI(&c)
			if err != nil {
				t.Error(err)
			}
			if err = compareUIPackets(tt.uiPackets, uiPackets); err != nil {
				t.Errorf("UI packets diff failed: %v", err)
			}

			// check packets received by each peer
			for i := 0; i < len(c.peers); i++ {
				pkts, err := readFromPeer(&c, i)
				if err != nil {
					t.Fatal(err)
				}
				if err = compareNetPackets(tt.peerPackets[i], pkts); err != nil {
					t.Fatalf("peer_%d diff failed: %v", i, err)
				}
			}
		})
	}

}

func compareUIPackets(expected []ui.Packet, got []ui.Packet) error {
	if len(expected) != len(got) {
		return fmt.Errorf("expected %d packages, got %d", len(expected), len(got))
	}
	for i, p := range got {
		if !reflect.DeepEqual(p, expected[i]) {
			return fmt.Errorf("expected packet %d to be %+v, got %+v", i, expected[i], p)
		}
	}
	return nil
}

func compareNetPackets(expected []Packet, got []Packet) error {
	if len(expected) != len(got) {
		return fmt.Errorf("expected %d packages, got %d", len(expected), len(got))
	}
	for i, p := range got {
		if !reflect.DeepEqual(p, expected[i]) {
			return fmt.Errorf("expected packet %d to be %+v, got %+v", i, expected[i], p)
		}
	}
	return nil
}
