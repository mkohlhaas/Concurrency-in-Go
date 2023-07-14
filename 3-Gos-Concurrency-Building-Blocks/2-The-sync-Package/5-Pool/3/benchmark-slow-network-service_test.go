package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

const (
	network = "tcp"
	url     = "localhost:8080"
)

func init() {
	daemonStarted := startNetworkDaemon()
	daemonStarted.Wait()
}

func connectToService() any {
	time.Sleep(1 * time.Second)
	return struct{}{}
}

func startNetworkDaemon() *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		server, err := net.Listen(network, url)
		if err != nil {
			log.Fatalf("cannot listen: %v", err)
		}
		defer server.Close()

		wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				log.Printf("cannot accept connection: %v", err)
				continue
			}
			connectToService()
			fmt.Fprintln(conn, "")
			conn.Close()
		}
	}()

	return &wg
}

func BenchmarkNetworkRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		conn, err := net.Dial(network, url)
		if err != nil {
			b.Fatalf("cannot dial host: %v", err)
		}
		if _, err := io.ReadAll(conn); err != nil {
			b.Fatalf("cannot read: %v", err)
		}
		conn.Close()
	}
}
