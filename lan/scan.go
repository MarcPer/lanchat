package lan

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os/exec"
)

type NetScanner interface {
	FindHost(int) (string, bool)
}

type DefaultScanner struct {
	Local bool
}

func (s *DefaultScanner) FindHost(port int) (string, bool) {
	ips := s.getIPs()
	return scanHost(ips, port)
}

func (s *DefaultScanner) getIPs() []net.IP {
	ips := make([]net.IP, 0)
	if s.Local {
		ips = append(ips, net.IPv4(127, 0, 0, 1))
	}

	cmd := `arp | awk '$1 ~ /^[0-9\.]+$/ {print $1}' | uniq`
	c := exec.Command("bash", "-c", cmd)
	var b bytes.Buffer
	c.Stdout = &b
	c.Run()

	scanner := bufio.NewScanner(&b)

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

func scanHost(ips []net.IP, chatPort int) (string, bool) {
	var found bool
	var url string
	for _, ip := range ips {
		url = fmt.Sprintf("%s:%d", ip, chatPort)
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
