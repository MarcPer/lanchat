package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const ChatPort = 6776
const serverUser = "server"
const notifyCooldown = 20 * time.Second

var clientsLock sync.RWMutex
var notifyLock sync.Mutex

const (
	fontReset  = "\033[0m"
	fontBold   = "\033[1m"
	fontGreen  = "\033[92m"
	fontBlue   = "\033[34m"
	fontYellow = "\033[93m"
)

type peer struct {
	id           string
	username     string
	conn         net.Conn
	inbox        chan packet
	outbox       chan packet
	server       bool
	clients      []*peer
	prompt       string
	promptDelete string
	lastNotify   time.Time
}

type packet struct {
	sourceUsr string
	sourceId  string
	destUsr   string
	msg       string
	hideNotif bool
}

func New(username string, server bool) peer {
	var cs []*peer
	if server {
		cs = make([]*peer, 0, 10)
	}
	inbox := make(chan packet, 50)
	outbox := make(chan packet, 50)
	prompt := fmt.Sprintf("%s%s%s>%s ", fontBold, fontGreen, username, fontReset)
	promptDelete := strings.Repeat("\b", len(prompt))
	return peer{
		username:     username,
		server:       server,
		clients:      cs,
		inbox:        inbox,
		outbox:       outbox,
		prompt:       prompt,
		promptDelete: promptDelete}
}

func (p *peer) Start(url string) {
	if p.server {
		p.StartServer()
	} else {
		p.StartClient(url)
		fmt.Print(p.prompt)
		p.outbox <- packet{sourceUsr: p.username, msg: fmt.Sprintf(":id %v", p.username)}
		p.outbox <- packet{sourceUsr: p.username, msg: ":info"}
		p.readInput()
	}
}

func (p *peer) StartServer() {
	url := fmt.Sprintf(":%d", ChatPort)
	ln, err := net.Listen("tcp", url)
	if err != nil {
		fmt.Printf("Could not start server: %v\n", err)
		os.Exit(1)
	}

	go p.pollConns()

	go p.processQueues()
	go p.readInput()
	fmt.Print(p.prompt)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v", err)
		} else {
			c := peer{id: conn.RemoteAddr().String(), conn: conn, inbox: p.inbox}
			p.clients = append(p.clients, &c)
			go c.readIncoming()
		}
	}
}

func (p *peer) StartClient(url string) {
	conn, err := net.Dial("tcp", url)
	p.conn = conn
	if err != nil {
		panic("Could not connect")
	}
	go p.readIncoming()
	go p.processQueues()
}

func (p *peer) pollConns() {
	for {
		clientsLock.RLock()
		for _, c := range p.clients {
			cli := c
			go func() {
				_, err := cli.conn.Write([]byte(":ping\n"))
				if err != nil {
					p.remove(cli.id)
				}
			}()
		}

		clientsLock.RUnlock()
		time.Sleep(1 * time.Second)
	}
}

func (p *peer) remove(clientID string) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	for idx, c := range p.clients {
		if c.id == clientID {
			p.clients[idx] = p.clients[len(p.clients)-1]
			p.clients = p.clients[:len(p.clients)-1]

			if c.username != "" {
				p.inbox <- packet{sourceUsr: serverUser, msg: c.username + " left"}
			}
			break
		}
	}
}

func (p *peer) readInput() {
	sca := bufio.NewScanner(os.Stdin)
	for sca.Scan() {
		p.outbox <- packet{sourceUsr: p.username, sourceId: p.id, msg: sca.Text()}
		fmt.Print(p.prompt)
	}
	if sca.Err() == nil {
		os.Exit(0)
	}
}

func (p *peer) readIncoming() {
	s := bufio.NewScanner(p.conn)
	connId := p.conn.RemoteAddr().String()
	for s.Scan() {
		data := strings.Split(s.Text(), "::")
		if len(data) > 1 {
			p.inbox <- packet{sourceId: connId, sourceUsr: data[0], msg: data[1]}
		}
	}
}

func (p *peer) processQueues() {
	for {
		select {
		case pk := <-p.inbox:
			if ok := p.processCmd(pk); ok {
				continue
			}
			if p.server {
				p.outbox <- pk
				if pk.destUsr != "" && pk.destUsr != p.username {
					continue
				}
			}
			go p.notify(pk)
			var color string
			fmt.Print(p.promptDelete, strings.Repeat(" ", len(p.promptDelete)), p.promptDelete)
			if pk.sourceUsr == serverUser {
				color = fontBlue
			} else {
				color = fontYellow
			}
			msg := fmt.Sprintf("%s%s%s>%s %s", fontBold, color, pk.sourceUsr, fontReset, pk.msg)
			fmt.Println(msg)
			fmt.Print(p.prompt)
		case pk := <-p.outbox:
			msg := fmt.Sprintf("%s::%s\n", pk.sourceUsr, pk.msg)
			if p.server {
				clientsLock.RLock()
				for _, c := range p.clients {
					if pk.destUsr != "" {
						if c.username == pk.destUsr {
							c.conn.Write([]byte(msg))
						}
					} else {
						if c.id != pk.sourceId {
							c.conn.Write([]byte(msg))
						}
					}
				}
				clientsLock.RUnlock()
			} else {
				p.conn.Write([]byte(msg))
			}
		}
	}
}

// Return true if command was successful
func (p *peer) processCmd(pk packet) bool {
	if !p.server {
		return false
	}
	if pk.msg == "" || pk.msg[0] != ':' {
		return false
	}
	str := strings.Split(pk.msg[1:], " ")
	var args []string
	cmd := str[0]
	if len(str) > 1 {
		args = str[1:]
	}

	switch cmd {
	case "id":
		if len(args) < 1 {
			goto FAIL_CMD
		} else {
			clientsLock.RLock()
			defer clientsLock.RUnlock()
			for _, c := range p.clients {
				if c.id == pk.sourceId {
					c.username = args[0]
					p.inbox <- packet{sourceId: c.id, sourceUsr: serverUser, msg: args[0] + " joined"}
				}
			}
			return true
		}
	case "info":
		p.sendInfoMsg(pk.sourceUsr)
		return true
	default:
		goto FAIL_CMD
	}
FAIL_CMD:
	fmt.Printf("Received invalid command: %v\n", pk.msg)
	return false
}

func (p *peer) sendInfoMsg(destUsr string) {
	p.inbox <- packet{hideNotif: true, sourceUsr: serverUser, destUsr: destUsr, msg: "Connected users:"}
	p.inbox <- packet{hideNotif: true, sourceUsr: serverUser, destUsr: destUsr, msg: "- " + p.username}
	clientsLock.RLock()
	defer clientsLock.RUnlock()
	for _, c := range p.clients {
		if c.username != "" {
			p.inbox <- packet{hideNotif: true, sourceUsr: serverUser, destUsr: destUsr, msg: "- " + c.username}
		}
	}
}

func (p *peer) notify(pk packet) {
	if pk.msg == "" || pk.hideNotif {
		return
	}
	notifyLock.Lock()
	defer notifyLock.Unlock()
	t := time.Now()
	if t.Sub(p.lastNotify) > notifyCooldown {
		p.lastNotify = t
		cmd := fmt.Sprintf("command -v notify-send && notify-send 'lanchat: %v> %v'", pk.sourceUsr, pk.msg)
		c := exec.Command("bash", "-c", cmd)
		c.Run()
	}
}
