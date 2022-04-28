package lan

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/MarcPer/lanchat/logger"
)

type NetScanner interface {
	FindHost(int) (string, bool)
}

type NullScanner struct{}

func (s *NullScanner) FindHost(port int) (string, bool) {
	return "", false
}

type DefaultScanner struct {
	Local bool
}

func (s *DefaultScanner) FindHost(port int) (string, bool) {
	targets, err := findTargets()
	if err != nil {
		logger.Errorf("FindHost: %v\n", err)
		return "", false
	}
	hosts := hostRange(targets)

	return s.scanHost(hosts, port)
}

type targetRange struct {
	selfIP string // IP from caller, to exclude from scan
	netIP  string // IPNet network in CIDR notation
}

// Returns list of target IP ranges to scan
func findTargets() ([]targetRange, error) {
	out := make([]targetRange, 0, 4)
	ifcs, err := net.Interfaces()
	if err != nil {
		return out, err
	}
	for _, ifc := range ifcs {
		addrs, err := ifc.Addrs()
		if err != nil {
			return out, err
		}
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					t := targetRange{ipnet.IP.String(), ipnet.String()}
					out = append(out, t)
					return out, nil
				}
			}
		}
	}

	return out, nil
}

func hostRange(targets []targetRange) []string {
	var hosts []string
	for _, t := range targets {
		_, ipv4Net, err := net.ParseCIDR(t.netIP)
		if err != nil {
			logger.Errorf("hostRange: %v\n", err)
			return []string{}
		}

		mask := binary.BigEndian.Uint32(ipv4Net.Mask)
		start := binary.BigEndian.Uint32(ipv4Net.IP)
		finish := (start & mask) | (mask ^ 0xffffffff)

		for i := start + 1; i <= finish-1; i++ {
			ip := make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, i)
			if ip.String() == t.selfIP {
				continue
			}
			hosts = append(hosts, ip.String())
		}

	}

	return hosts
}

func (s *DefaultScanner) scanHost(ips []string, chatPort int) (string, bool) {
	if s.Local {
		url := fmt.Sprintf("%s:%d", "127.0.0.1", chatPort)
		logger.Debugf("scanning %s\n", url)
		conn, err := net.DialTimeout("tcp", url, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return url, true
		}
	}

	if len(ips) < 1 {
		logger.Infof("No hosts to scan\n")
		return "", false
	}

	hostInCh := make(chan string)
	logger.Debugf("Scanning %d hosts\n", len(ips))
	go func() {
		for _, host := range ips {
			hostInCh <- host
		}
		close(hostInCh)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	resCh := make(chan string)
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go popScan(ctx, chatPort, hostInCh, resCh)
	}

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true

	}()

	select {
	case url := <-resCh:
		cancel()
		return url, true
	case <-done:
		cancel()
		return "", false
	}
}

var wg sync.WaitGroup

const numWorkers = 10
const timeout = 100 * time.Millisecond

func popScan(ctx context.Context, chatPort int, in chan string, out chan string) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case host, ok := <-in:
			if !ok {
				wg.Done()
				return
			}
			url := fmt.Sprintf("%s:%d", host, chatPort)
			logger.Debugf("scanning %s", url)
			conn, err := net.DialTimeout("tcp", url, timeout)
			if err == nil {
				conn.Close()
				out <- url
				wg.Done()
				return
			}
		}
	}
}
