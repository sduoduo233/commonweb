package main

import (
	c "commonweb/client"
	s "commonweb/server"
	"flag"
	"log"
)

var server = flag.Bool("server", false, "server mode")

var local = flag.String("local", "0.0.0.0:8081", "local address")
var remote = flag.String("remote", "127.0.0.1:8080", "remote address")
var upPath = flag.String("up", "/up", "upload path")
var downPath = flag.String("down", "/down", "download path")
var upUrl = flag.String("upUrl", "http://127.0.0.1:8081/up", "upload url")
var downUrl = flag.String("downUrl", "http://127.0.0.1:8081/down", "upload url")

func main() {
	flag.Parse()

	log.Printf("commonweb")

	if *server {
		s.StartServer(s.ServerConfig{
			Listen:   *local,
			Remote:   *remote,
			UpPath:   *upPath,
			DownPath: *downPath,
		})
	} else {
		c.StartClient(c.ClientConfig{
			Listen:  *local,
			UpURL:   *upUrl,
			DownURL: *downUrl,
		})
	}
}
