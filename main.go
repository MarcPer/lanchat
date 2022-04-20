package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/MarcPer/lanchat/lan"
	"github.com/MarcPer/lanchat/logger"
	"github.com/MarcPer/lanchat/ui"
)

func main() {
	var user = flag.String("u", "noone", "user name")
	var local = flag.Bool("local", true, "whether to search for a running server in localhost")
	var notify = flag.Bool("notify", true, "whether to send system notifications upon message receivals. Notifications have a cooldown time.")
	var port = flag.Int("p", 6776, "port ")
	flag.Parse()
	ui.EnableNotification = *notify
	toUI := make(chan ui.Packet, 2)    // used by client to send info to UI
	fromUI := make(chan ui.Packet, 10) // used by UI to send info to client
	client := &lan.Client{Name: *user, HostPort: *port, FromUI: fromUI, ToUI: toUI, Scanner: &lan.DefaultScanner{Local: *local}}

	ctx, cancel := context.WithCancel(context.Background())
	client.Start(ctx)
	logger.Infof("Starting UI\n")
	renderer := ui.New(*user, toUI, fromUI)
	logger.Init(&renderer)
	renderer.Run()

	cancel()
	fmt.Println("bye")
}
