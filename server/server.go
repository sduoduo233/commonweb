package server

import (
	"io"
	"log"
	"net"
	"net/http"
	"sync"
)

var config ServerConfig

type ServerConfig struct {
	Listen   string
	Remote   string
	UpPath   string
	DownPath string
}

type Session struct {
	conn net.Conn
}

func newSession() (*Session, error) {
	conn, err := net.Dial("tcp", config.Remote)
	if err != nil {
		return nil, err
	}

	return &Session{
		conn: conn,
	}, nil
}

func StartServer(cfg ServerConfig) {
	config = cfg

	var sessions sync.Map

	http.HandleFunc(config.DownPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		sessionId := r.URL.Query().Get("session")
		if sessionId == "" {
			io.WriteString(w, "invalid session")
			return
		}
		log.Printf("new down connection: %s", sessionId)
		defer log.Printf("down connection end: %s", sessionId)

		s, ok := sessions.Load(sessionId)

		var session *Session
		if !ok {
			// create a new connection

			var err error
			session, err = newSession()
			if err != nil {
				log.Printf("dial error: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer session.conn.Close()
			defer sessions.Delete(sessionId)

			sessions.Store(sessionId, session)
		} else {
			session = s.(*Session)
		}

		// forward response to client
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)

		buf := make([]byte, 4096)
		for {
			n, err := session.conn.Read(buf)
			if err != nil {
				log.Printf("read down error: %s", err)
				return
			}
			_, err = w.Write(buf[:n])
			if err != nil {
				log.Printf("write down error: %s", err)
				return
			}
			w.(http.Flusher).Flush()
		}
	})

	http.HandleFunc(config.UpPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		sessionId := r.URL.Query().Get("session")
		if sessionId == "" {
			io.WriteString(w, "invalid session")
			return
		}
		log.Printf("new up connection: %s", sessionId)
		defer log.Printf("up connection end: %s", sessionId)

		s, ok := sessions.Load(sessionId)

		var session *Session
		if !ok {
			// create a new connection

			var err error
			session, err = newSession()
			if err != nil {
				log.Printf("dial error: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer session.conn.Close()
			defer sessions.Delete(sessionId)

			sessions.Store(sessionId, session)
		} else {
			session = s.(*Session)
		}

		// forward request to server
		buf := make([]byte, 4096)
		for {
			eof := false

			n, err := r.Body.Read(buf)
			if err != nil {
				if err == io.EOF {
					eof = true
				} else {
					log.Printf("read up error: %s", err)
					w.WriteHeader(http.StatusOK)
					return
				}
			}

			_, err = session.conn.Write(buf[:n])
			if err != nil {
				log.Printf("write up error: %s", err)
				w.WriteHeader(http.StatusOK)
				return
			}

			if eof {
				return
			}
		}
	})

	err := http.ListenAndServe(config.Listen, nil)
	if err != nil {
		log.Fatalf("serve: %s", err)
	}
}
