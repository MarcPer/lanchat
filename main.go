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
	fmt.Printf("%+v\n", cfg)
	ui.EnableNotification = cfg.notify
	toUI := make(chan ui.Packet, 2)    // used by client to send info to UI
	fromUI := make(chan ui.Packet, 10) // used by UI to send info to client
	client := &lan.Client{Name: cfg.username, HostPort: cfg.port, FromUI: fromUI, ToUI: toUI, Scanner: &lan.DefaultScanner{Local: cfg.local}}

	ctx, cancel := context.WithCancel(context.Background())
	client.Start(ctx)
	logger.Infof("Starting UI\n")
	renderer := ui.New(cfg.username, toUI, fromUI)
	logger.Init(&renderer)
	renderer.Run()

	cancel()
	fmt.Println("bye")
}
