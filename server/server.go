package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	lGreen  = "\033[92m"
	lYellow = "\033[93m"
)

type client struct {
	id         string
	conn       net.Conn
	sndChannel chan []byte
	rcvChannel chan packet
}

type Server struct {
	id           string
	clients      []*client
	rcvChannel   chan packet
	prompt       string
	promptDelete string
}

type packet struct {
	id  string
	msg string
}

func New(username string) Server {
	cs := make([]*client, 0, 10)
	ch := make(chan packet, 50)
	prompt := fmt.Sprintf("%s%s%s>%s ", bold, lGreen, username, string(reset))
	promptDelete := strings.Repeat("\b", len(prompt))
	return Server{username, cs, ch, prompt, promptDelete}
}

func (s *Server) Start(url string) {
	ln, err := net.Listen("tcp", url)
	if err != nil {
		fmt.Printf("Could not start server: %v\n", err)
		os.Exit(1)
	}

	go s.startBroadcast()
	go s.readInput()
	// go s.pollConns()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v", err)
		} else {
			sndCh := make(chan []byte, 50)
			c := client{conn: conn, sndChannel: sndCh, rcvChannel: s.rcvChannel}
			s.clients = append(s.clients, &c)
			go handleConn(&c)
		}
	}
}

func handleConn(c *client) {
	go c.readIncoming()

	for msg := range c.sndChannel {
		c.conn.Write(msg)
	}
}

func (s *Server) pollConns() {
	for {
		for _, c := range s.clients {
			cli := c
			go func() {
				cli.conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
				_, err := cli.conn.Write([]byte(":ping\n"))
				if err != nil {
					fmt.Fprintf(os.Stdout, "Oh no! %v", err)
					s.remove(cli.id)
				}
			}()
		}

		time.Sleep(8 * time.Second)
	}
}

func (s *Server) remove(clientID string) {
	var mtx sync.Mutex
	mtx.Lock()
	defer mtx.Unlock()

	for idx, c := range s.clients {
		if c.id == clientID {
			s.clients[idx] = s.clients[len(s.clients)-1]
			s.clients = s.clients[:len(s.clients)-1]

			s.rcvChannel <- packet{s.id, fmt.Sprintf("%v left\n", clientID)}
			break
		}
	}
}

func (s *Server) readInput() {
	sca := bufio.NewScanner(os.Stdin)
	for sca.Scan() {
		s.rcvChannel <- packet{s.id, sca.Text()}
		fmt.Print(s.prompt)
	}
}

func (c *client) readIncoming() {
	s := bufio.NewScanner(c.conn)
	for s.Scan() {
		msg := s.Text()
		if msg != "" && msg[0] == ':' {
			processCmd(c, msg[1:])
		} else {
			c.rcvChannel <- packet{c.id, s.Text()}
		}
	}
}

func (s *Server) startBroadcast() {
	for pk := range s.rcvChannel {
		if pk.id != s.id {
			fmt.Print(s.promptDelete)
			msg := fmt.Sprintf("%s%s%s>%s %s", string(bold), string(lYellow), pk.id, string(reset), pk.msg)
			fmt.Println(msg)
			fmt.Print(s.prompt)
		}

		for _, c := range s.clients {
			if c.id != pk.id {
				packet := fmt.Sprintf("%s::%s\n", pk.id, pk.msg)
				c.conn.Write([]byte(packet))
			}
		}
	}
}

func processCmd(c *client, msg string) {
	str := strings.Split(msg, " ")
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
			if c.id == "" {
				c.id = args[0]
			}
			return
		}
	default:
		goto FAIL_CMD
	}
FAIL_CMD:
	fmt.Printf("Received invalid command: %v\n", msg)
}
