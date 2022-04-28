package main

import (
	"context"
	"fmt"
	"log"
	"os"

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
	logger.Infof("Starting UI\n")
	renderer := ui.New(cfg.username, toUI, fromUI)
	logger.Init(&renderer)
	f := debugFile()
	defer f.Close()
	logger.InitDebug(f)

	client := &lan.Client{Name: cfg.username, HostPort: cfg.port, FromUI: fromUI, ToUI: toUI, Scanner: scanner}
	ctx, cancel := context.WithCancel(context.Background())
	client.Start(ctx)
	renderer.Run()

	cancel()
	fmt.Println("bye")
}

func debugFile() *os.File {
	var debugPath string
	if logger.LogLevel >= logger.LogLevelDebug {
		debugPath = "debug.log"
	} else {
		debugPath = os.DevNull
	}

	f, err := os.OpenFile(debugPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return f
}
