package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	lGreen  = "\033[92m"
	lYellow = "\033[93m"
)

type Client struct {
	id        string
	serverURL string
}

type packet struct {
	id  string
	msg string
}

func New(username string, servURL string) Client {
	return Client{username, servURL}
}

func (c *Client) Start() {
	conn, err := net.Dial("tcp", c.serverURL)
	if err != nil {
		panic("Could not connect")
	}
	prompt := fmt.Sprintf("%s%s%s>%s ", string(bold), string(lGreen), c.id, string(reset))
	promptDelete := strings.Repeat("\b", len(prompt))
	sendCh := make(chan []byte, 50)
	msg := fmt.Sprintf(":id %v\n", c.id)
	conn.Write([]byte(msg))
	conn.Write([]byte(":info\n"))

	fmt.Print(prompt)
	go readInput(sendCh)

	rcvCh := make(chan packet, 10)
	go c.readIncoming(rcvCh, conn)

	ping := ":ping\n"
	for {
		select {
		case pk := <-sendCh:
			fmt.Print(prompt)
			conn.Write(pk)
		case pk := <-rcvCh:
			if pk.msg != ping {
				fmt.Print(promptDelete)
				msg := fmt.Sprintf("%s%s%s>%s %s", string(bold), string(lYellow), pk.id, string(reset), pk.msg)
				fmt.Println(msg)
				fmt.Print(prompt)
			}
		}
	}
}

func readInput(sendCh chan []byte) {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		msg := fmt.Sprintf("%v\n", s.Text())
		sendCh <- []byte(msg)
	}
}

func (c *Client) readIncoming(rcvCh chan packet, conn net.Conn) {
	s := bufio.NewScanner(conn)
	for s.Scan() {
		data := strings.Split(s.Text(), "::")
		rcvCh <- packet{data[0], data[1]}
	}
}
