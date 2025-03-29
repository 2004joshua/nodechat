package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/2004joshua/nodechat/internal/db"
	"github.com/2004joshua/nodechat/internal/model"
	"github.com/2004joshua/nodechat/internal/peer"
)

func main() {
	port := flag.String("port", "", "port to listen on")
	connect := flag.String("connect", "", "peer to connect to (ip:port)")
	username := flag.String("username", "", "username for this peer")
	flag.Parse()

	if *port == "" || *username == "" {
		fmt.Println("Usage: go run main.go --port=PORT --username=NAME [--connect=IP:PORT]")
		os.Exit(1)
	}

	// Create the databases directory if it doesn't exist
	dbDir := "databases"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		panic(err)
	}

	// Build a database path unique to the username.
	dbPath := filepath.Join(dbDir, *username+".db")

	// Initialize the user's SQLite database.
	if err := db.InitDB(dbPath); err != nil {
		panic(err)
	}

	// Create and run the peer instance.
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
			Sender:  *username,
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
