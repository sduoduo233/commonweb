package client

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

type ClientConfig struct {
	Listen  string
	UpURL   string
	DownURL string
}

var config ClientConfig

func StartClient(cfg ClientConfig) {
	config = cfg

	server, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		log.Fatalf("listen: %s", err)
	}

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatalf("accept: %s", err)
		}

		go handleConn(conn)
	}

}

func handleConn(conn net.Conn) {
	defer conn.Close()

	session := uuid.New().String()
	log.Printf("new connection: %s", session)
	defer log.Printf("connection end: %s", session)

	var wg sync.WaitGroup
	wg.Add(2)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer wg.Done()
		defer cancel()

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s?session=%s", config.UpURL, session), conn)
		req = req.WithContext(ctx)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			cancel()
			log.Printf("post error: %s", err)
			return
		}

		resp.Body.Close()

	}()

	go func() {
		defer wg.Done()
		defer cancel()

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s?session=%s", config.DownURL, session), nil)
		req = req.WithContext(ctx)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("get error: %s", err)
			return
		}

		buf := make([]byte, 4096)

		for {
			n, err := resp.Body.Read(buf)
			if err != nil {
				log.Printf("get read error: %s", err)
				return
			}

			_, err = conn.Write(buf[:n])
			if err != nil {
				log.Printf("get write error: %s", err)
				return
			}
		}
	}()

	wg.Wait()
}
