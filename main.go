package main

import (
	"context"
	"fmt"

	"github.com/MarcPer/lanchat/lan"
	"github.com/MarcPer/lanchat/logger"
	"github.com/MarcPer/lanchat/ui"
)

func main() {
	cfg := newConfig()
	ui.EnableNotification = cfg.notify
	toUI := make(chan ui.Packet, 2)    // used by client to send info to UI
	fromUI := make(chan ui.Packet, 10) // used by UI to send info to client

	var scanner lan.NetScanner
	if cfg.forceHost {
		scanner = &lan.NullScanner{}
	} else {
		scanner = &lan.DefaultScanner{Local: cfg.local}

	}
	client := &lan.Client{Name: cfg.username, HostPort: cfg.port, FromUI: fromUI, ToUI: toUI, Scanner: scanner}
	ctx, cancel := context.WithCancel(context.Background())
	client.Start(ctx)
	logger.Infof("Starting UI\n")
	renderer := ui.New(cfg.username, toUI, fromUI)
	logger.Init(&renderer)
	renderer.Run()

	cancel()
	fmt.Println("bye")
}
