package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/2004joshua/nodechat/internal/db"
	"github.com/2004joshua/nodechat/internal/model"
	"github.com/2004joshua/nodechat/internal/peer"
)

func main() {
	port := flag.String("port", "", "port to listen on for P2P")
	connect := flag.String("connect", "", "peer to connect to (ip:port)")
	username := flag.String("username", "", "username for this peer")
	apiPort := flag.String("api-port", "8080", "port for API endpoints")
	flag.Parse()

	if *port == "" || *username == "" {
		fmt.Println("Usage: go run main.go --port=PORT --username=NAME [--connect=IP:PORT] [--api-port=PORT]")
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
	p := peer.New(":"+*port, *username)
	if err := p.Listen(); err != nil {
		panic(err)
	}

	if *connect != "" {
		if err := p.Connect(*connect); err != nil {
			panic(err)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only redirect if no query params or missing username
		if r.URL.Path == "/" && r.URL.Query().Get("username") == "" {
			http.Redirect(w, r, "/?username="+*username, http.StatusFound)
			return
		}
		// Serve static files (React build)
		http.FileServer(http.Dir("ui/build")).ServeHTTP(w, r)
	})

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		if topic == "" {
			http.Error(w, "Missing topic", http.StatusBadRequest)
			return
		}
		p.Subscribe(topic)
		if err := db.SaveSubscription(*username, topic); err != nil {
			fmt.Println("Error saving subscription:", err)
		}
		w.Write([]byte("Subscribed to " + topic))
	})

	http.HandleFunc("/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		if topic == "" {
			http.Error(w, "Missing topic", http.StatusBadRequest)
			return
		}
		p.Unsubscribe(topic)
		if err := db.RemoveSubscription(*username, topic); err != nil {
			fmt.Println("Error removing subscription:", err)
		}

		w.Write([]byte("Unsubscribed from " + topic))
	})

	http.HandleFunc("/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		topics, err := db.GetSubscriptions(*username)
		if err != nil {
			http.Error(w, "Failed to retrieve subscriptions", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topics)
	})

	// Start API server in a separate goroutine.
	go func() {
		// GET /messages returns all stored messages.
		http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				messages, err := db.GetMessages()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(messages)
			} else if r.Method == http.MethodPost {
				// POST /messages accepts a JSON message, stores it, and broadcasts it.
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Unable to read request body", http.StatusBadRequest)
					return
				}
				defer r.Body.Close()

				var msg model.Message
				if err := json.Unmarshal(body, &msg); err != nil {
					http.Error(w, "Invalid JSON", http.StatusBadRequest)
					return
				}

				// Set sender if not provided (optional: override if needed)
				if msg.Sender == "" {
					msg.Sender = *username
				}

				// Store and broadcast the message.
				encoded, err := msg.Encode()
				if err != nil {
					http.Error(w, "Error encoding message", http.StatusInternalServerError)
					return
				}
				// Broadcast to connected peers.
				p.Broadcast(encoded)

				// Save to local database.
				if err := db.SaveMessage(&msg); err != nil {
					http.Error(w, "Error saving message", http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("Message received"))
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		fmt.Printf("API listening on port %s...\n", *apiPort)
		if err := http.ListenAndServe(":"+*apiPort, nil); err != nil {
			panic(err)
		}
	}()

	// Read from stdin and broadcast using our JSON protocol.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		if len(input) >= 10 && input[:10] == "/subscribe" {
			topic := strings.TrimSpace(input[11:])
			p.Subscribe(topic)
			continue
		}

		if len(input) >= 12 && input[:12] == "/unsubscribe" {
			topic := strings.TrimSpace(input[13:])
			p.Unsubscribe(topic)
			continue
		}

		if input == "/exit" {
			fmt.Printf("Goodbye %s...\n", *username)
			os.Exit(0)
		}

		if input == "/help" {
			fmt.Println("Available commands:")
			fmt.Println("/subscribe <topic> - Subscribe to a topic")
			fmt.Println("/unsubscribe <topic> - Unsubscribe from a topic")
			fmt.Println("/exit - Exit the chat")
			fmt.Println("To send a topic-specific message, use: /topic <topic> <message>")
			continue
		}

		var topic string
		var content string

		if strings.HasPrefix(input, "/topic ") {
			parts := strings.SplitN(input, " ", 3)
			if len(parts) < 3 {
				fmt.Println("Usage: /topic <topic> <message>")
				continue
			}
			topic = parts[1]
			content = parts[2]
		} else {
			content = input
		}

		msg := &model.Message{
			Type:    "chat",
			Sender:  *username,
			Content: content,
			Topic:   topic,
		}

		encodedMsg, err := msg.Encode()
		if err != nil {
			fmt.Println("Error encoding message:", err)
			continue
		}
		p.Broadcast(encodedMsg)
	}
}
