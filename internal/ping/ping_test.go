package ping

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestPingHost(t *testing.T) {
	// open listing socket
	cancelCtx, cfun := context.WithCancel(context.Background())
	defer cfun()
	listenInstance := net.ListenConfig{}
	listener, err := listenInstance.Listen(cancelCtx, "tcp", "127.0.0.1:")

	if err != nil {
		t.Fatalf("Failed to create tcp socket to listen on random port, ERROR: %v", err)
	}
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to get port from Addr string. Error: %v", err)
	}
	t.Logf("Listening on Addr %v", listener.Addr())
	portNum, _ := strconv.Atoi(port)
	res := PingRemoteHost("127.0.0.1", uint(portNum), time.Second*1)
	if res.Err != nil {
		t.Fatalf("Received error ping valid host. Error %v", res.Err)
	}
	if !res.Reachable {
		t.Fatalf("Should be able to reach local host")
	}
	listener.Close()
}

func TestPingHostConnectionRefused(t *testing.T) {
	// dial localhost with a unused port
	res := PingRemoteHost("localhost", 5673, time.Millisecond*150)
	if res.Reachable {
		t.Fatalf("Should not be able to reach host destination")
	}
	if res.Err != nil {
		t.Fatalf("there should not be a reported error after connection refused message: Error %v", res.Err)
	}
	t.Logf("Reported connection latency is %v", res.Latency)
}

func TestPingHostUnreachable(t *testing.T) {
	notRealHost := "192.168.60.128"
	res := PingRemoteHost(notRealHost, 22, time.Millisecond*250)
	if res.Err == nil {
		t.Fatalf("Should have received a host not found error but received no error")
	}
	if res.Reachable {
		t.Fatalf("Host should not be considered reachable")
	}
	t.Logf("Error Received pinging non existent host %v", res.Err)
}
