package main

import (
	"context"
	"flag"

	"github.com/MarcPer/lanchat/lan"
	"github.com/MarcPer/lanchat/ui"
)

func main() {
	var user = flag.String("u", "noone", "user name")
	var local = flag.Bool("local", false, "whether to search for a running server in localhost")
	var port = flag.Int("p", 6776, "port ")
	flag.Parse()
	toUi := make(chan ui.Packet, 2)    // used by client to send info to UI
	fromUi := make(chan ui.Packet, 10) // used by UI to send info to client
	client := &lan.Client{Name: *user, Local: *local, HostPort: *port, FromUI: fromUi, ToUI: toUi, Scanner: &lan.DefaultScanner{Local: *local}}

	ctx := context.Background()
	client.Start(ctx)
}
