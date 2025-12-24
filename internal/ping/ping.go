package ping

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"syscall"
	"time"
)

type PingResult struct {
	Reachable bool
	Latency   time.Duration
	Err       error
}

func getIpFromHostname(hostname string) (net.IP, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}
	return ips[0], nil
}

func PingRemoteHost(hostname string, port uint, timeout time.Duration) PingResult {
	ip, err := getIpFromHostname(hostname)
	if err != nil {
		return PingResult{
			Reachable: false,
			Err:       err,
		}
	}
	dialer := net.Dialer{
		Timeout: timeout,
	}
	addr := net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
	start := time.Now()
	conn, err := dialer.Dial("tcp", addr)
	rtt := time.Since(start)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return PingResult{
				Reachable: false,
				Latency:   rtt,
				Err:       nil,
			}

		} else {
			return PingResult{
				Reachable: false,
				Err:       fmt.Errorf("Failed to ping host, Error: %v", err),
			}
		}
	}
	conn.Close()
	return PingResult{
		Reachable: true,
		Latency:   rtt,
		Err:       nil,
	}
}
