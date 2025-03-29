package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/2004joshua/nodechat/internal/db"
	"github.com/2004joshua/nodechat/internal/model"
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

	// Initialize SQLite database
	if err := db.InitDB("messages.db"); err != nil {
		panic(err)
	}

	p := peer.New(":" + *port)

	if err := p.Listen(); err != nil {
		panic(err)
	}

	if *connect != "" {
		if err := p.Connect(*connect); err != nil {
			panic(err)
		}
	}

	// Read from stdin and broadcast using our JSON protocol.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msgText := scanner.Text()
		// Create a chat message using our JSON-based protocol.
		msg := &model.Message{
			Type:    "chat",
			Sender:  "self", // Replace with actual user identity if needed.
			Content: msgText,
		}
		encodedMsg, err := msg.Encode()
		if err != nil {
			fmt.Println("Error encoding message:", err)
			continue
		}
		p.Broadcast(encodedMsg)
	}
}
