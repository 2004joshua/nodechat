package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/2004joshua/nodechat/internal/peer"
	"github.com/2004joshua/nodechat/internal/storage"
)

func main() {
	port := flag.String("port", "", "port to listen on")
	connect := flag.String("connect", "", "peer to connect to (ip:port)")
	name := flag.String("name", "local", "your peer name")
	dbFile := flag.String("db", "peer.db", "SQLite database file name (will be stored in the databases directory)")
	flag.Parse()

	if *port == "" {
		fmt.Println("Usage: go run main.go --port=PORT [--connect=IP:PORT] [--name=YourName] [--db=YourDBFile]")
		os.Exit(1)
	}

	// Ensure the "databases" directory exists in the project root.
	dbDir := "databases"
	err := os.MkdirAll(dbDir, 0755)
	if err != nil {
		panic("Failed to create databases directory: " + err.Error())
	}

	// Construct the full path for the database file.
	dbPath := filepath.Join(dbDir, *dbFile)

	// Initialize the SQLite database.
	db, err := storage.NewDB(dbPath)
	if err != nil {
		panic("Database initialization failed: " + err.Error())
	}

	// Create a new peer with the specified port, name, and database.
	p := peer.New(":"+*port, *name, db)

	// Start listening for incoming connections.
	err = p.Listen()
	if err != nil {
		panic(err)
	}

	// If a peer to connect to is specified, connect to it.
	if *connect != "" {
		err := p.Connect(*connect)
		if err != nil {
			panic(err)
		}
	}

	// Read messages from stdin and broadcast them.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		p.Broadcast(msg)
	}
}
