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
	url         = "localhost:8080"
	network     = "tcp"
	numServices = 10
)

func init() {
	daemonStarted := startNetworkDaemon()
	daemonStarted.Wait()
}

func startNetworkDaemon() *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		connPool := warmServiceConnCache()

		server, err := net.Listen(network, url)
		if err != nil {
			log.Fatalf("cannot listen: %v", err)
		}
		defer server.Close()

		wg.Done() // let go of the init function

		// server loop
		for {
			// Only one connection at a time possible.
			// Normally it would run in a separate goroutine.
			conn, err := server.Accept()
			if err != nil {
				log.Printf("cannot accept connection: %v", err)
				continue
			}
			svcConn := connPool.Get()
			fmt.Fprintln(conn, "")
			connPool.Put(svcConn)
			conn.Close()
		}
	}()

	return &wg
}

func warmServiceConnCache() *sync.Pool {
	p := &sync.Pool{
		New: connectToService,
	}
	for i := 0; i < numServices; i++ {
		p.Put(p.New())
	}
	return p
}

// create a dummy service
func connectToService() any {
	time.Sleep(1 * time.Second) // service is expensive to set up
	return struct{}{}
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
