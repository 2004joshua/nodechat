package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/2004joshua/nodechat/internal/peer"
)

func main() {
	port := flag.String("port", "", "port to listen on")
	connect := flag.String("connect", "", "peer to connect to (ip:port)")
	flag.Parse()

	if *port == "" {
		fmt.Println("Usage: go run main.go --port=PORT [--connect=IP:PORT]")
		os.Exit(1)
	}

	p := peer.New(":" + *port)

	err := p.Listen()
	if err != nil {
		panic(err)
	}

	if *connect != "" {
		err := p.Connect(*connect)
		if err != nil {
			panic(err)
		}
	}

	// Broadcast messages from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		p.Broadcast(msg)
	}
}
