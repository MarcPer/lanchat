package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os/exec"

	"github.com/MarcPer/lanchat/client"
	"github.com/MarcPer/lanchat/server"
)

const CHAT_PORT = 6776

func main() {
	var user = flag.String("u", "noone", "user name")
	var local = flag.Bool("local", false, "whether to search for a running server in localhost")
	flag.Parse()
	ips := getIps()

	if *local {
		ips = append(ips, net.IPv4(127, 0, 0, 1))
	}
	url, found := searchServer(ips)

	if found {
		fmt.Printf("found in %s\n", url)
		c := client.New(*user, url)
		c.Start()
	} else {
		fmt.Println("no active servers found. Starting one.")
		s := server.New(*user)
		s.Start(fmt.Sprintf(":%d", CHAT_PORT))
	}
}

func getIps() []net.IP {
	cmd := `arp | awk '$1 ~ /^[0-9\.]+$/ {print $1}' | uniq`
	c := exec.Command("bash", "-c", cmd)
	var b bytes.Buffer
	c.Stdout = &b
	c.Run()

	scanner := bufio.NewScanner(&b)

	ips := make([]net.IP, 0, 8)
	for scanner.Scan() {
		rawIp := scanner.Text()
		ip := net.ParseIP(rawIp)
		if ip != nil {
			if !ip.IsLoopback() && ip[len(ip)-1] != 255 {
				ips = append(ips, ip)
			}
		}
	}

	return ips
}

func searchServer(ips []net.IP) (string, bool) {
	found := false
	fmt.Print("Searching for active server... ")
	var url string
	for _, ip := range ips {
		url = fmt.Sprintf("%s:%d", ip, CHAT_PORT)
		conn, err := net.Dial("tcp", url)
		if err == nil {
			conn.Close()
			found = true
			break
		}
	}

	if found {
		return url, found
	} else {
		return "", found
	}
}
