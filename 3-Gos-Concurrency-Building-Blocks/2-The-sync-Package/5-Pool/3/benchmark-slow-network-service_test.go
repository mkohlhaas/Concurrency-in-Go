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
	time.Sleep(time.Second)
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
		defer func() {
			if err := server.Close(); err != nil {
				log.Printf("cannot close server: %v", err)
			}
		}()

		wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				log.Printf("cannot accept connection: %v", err)
				continue
			}
			connectToService()
			if _, err := fmt.Fprintln(conn, ""); err != nil {
				log.Printf("cannot write response: %v", err)
			}
			if err := conn.Close(); err != nil {
				log.Printf("cannot close connection: %v", err)
			}
		}
	}()

	return &wg
}

func BenchmarkNetworkRequest(b *testing.B) {
	for b.Loop() {
		conn, err := net.Dial(network, url)
		if err != nil {
			b.Fatalf("cannot dial host: %v", err)
		}
		if _, err := io.ReadAll(conn); err != nil {
			b.Fatalf("cannot read: %v", err)
		}
		if err := conn.Close(); err != nil {
			b.Fatalf("cannot close connection: %v", err)
		}
	}
}
