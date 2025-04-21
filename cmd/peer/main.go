package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/2004joshua/nodechat/internal/db"
	"github.com/2004joshua/nodechat/internal/model"
	"github.com/2004joshua/nodechat/internal/peer"
	"github.com/google/uuid"
)

func main() {
	// CLI flags
	port := flag.String("port", "", "port to listen on for P2P")
	connect := flag.String("connect", "", "peer to connect to (ip:port)")
	username := flag.String("username", "", "username for this peer")
	apiPort := flag.String("api-port", "8080", "port for API endpoints")
	flag.Parse()

	if *port == "" || *username == "" {
		fmt.Println("Usage: go run main.go --port=PORT --username=NAME [--connect=IP:PORT] [--api-port=PORT]")
		os.Exit(1)
	}

	// Ensure database directory
	dbDir := "databases"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		panic(err)
	}
	dbPath := filepath.Join(dbDir, *username+".db")
	if err := db.InitDB(dbPath); err != nil {
		panic(err)
	}

	// Start P2P listener
	p := peer.New(":"+*port, *username)
	if err := p.Listen(); err != nil {
		panic(err)
	}
	if *connect != "" {
		if err := p.Connect(*connect); err != nil {
			panic(err)
		}
	}

	// 1) Serve React UI & subscriptions
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Redirect bare "/" to include ?username=
		if r.URL.Path == "/" && r.URL.Query().Get("username") == "" {
			http.Redirect(w, r, "/?username="+*username, http.StatusFound)
			return
		}
		// Serve files from ui/build
		http.FileServer(http.Dir("ui/build")).ServeHTTP(w, r)
	})

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		t := r.URL.Query().Get("topic")
		if t == "" {
			http.Error(w, "Missing topic", http.StatusBadRequest)
			return
		}
		p.Subscribe(t)
		db.SaveSubscription(*username, t)
		w.Write([]byte("Subscribed to " + t))
	})

	http.HandleFunc("/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		t := r.URL.Query().Get("topic")
		if t == "" {
			http.Error(w, "Missing topic", http.StatusBadRequest)
			return
		}
		p.Unsubscribe(t)
		db.RemoveSubscription(*username, t)
		w.Write([]byte("Unsubscribed from " + t))
	})

	http.HandleFunc("/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		topics, _ := db.GetSubscriptions(*username)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topics)
	})

	// 2) Prepare uploads directory
	os.MkdirAll("uploads", 0755)

	// 3) GIF upload endpoint
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(5 << 20); err != nil {
			http.Error(w, "file too big", http.StatusRequestEntityTooLarge)
			return
		}
		f, hdr, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "invalid upload", http.StatusBadRequest)
			return
		}
		defer f.Close()

		if hdr.Header.Get("Content-Type") != "image/gif" {
			http.Error(w, "only GIFs allowed", http.StatusBadRequest)
			return
		}

		filename := uuid.New().String() + filepath.Ext(hdr.Filename)
		dst, err := os.Create(filepath.Join("uploads", filename))
		if err != nil {
			http.Error(w, "cannot save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		io.Copy(dst, f)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"url":      "/uploads/" + filename,
			"fileName": hdr.Filename,
		})
	})

	// 4) Serve uploaded GIFs under /uploads/
	http.Handle(
		"/uploads/",
		http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))),
	)

	// 5) Start the API for /messages
	go func() {
		http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				msgs, _ := db.GetMessages()
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(msgs)

			case http.MethodPost:
				b, _ := ioutil.ReadAll(r.Body)
				defer r.Body.Close()

				var msg model.Message
				if err := json.Unmarshal(b, &msg); err != nil {
					http.Error(w, "Invalid JSON", http.StatusBadRequest)
					return
				}
				if msg.Sender == "" {
					msg.Sender = *username
				}
				if enc, err := msg.Encode(); err == nil {
					p.Broadcast(enc)
				}
				db.SaveMessage(&msg)

				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("Message received"))

			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		fmt.Printf("API listening on port %s...\n", *apiPort)
		if err := http.ListenAndServe(":"+*apiPort, nil); err != nil {
			panic(err)
		}
	}()

	// 6) CLI loop for topic commands and chat
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "/subscribe "):
				p.Subscribe(line[len("/subscribe "):])
			case strings.HasPrefix(line, "/unsubscribe "):
				p.Unsubscribe(line[len("/unsubscribe "):])
			case strings.HasPrefix(line, "/topic "):
				parts := strings.SplitN(line, " ", 3)
				if len(parts) == 3 {
					msg := &model.Message{
						Type:    "chat",
						Sender:  *username,
						Content: parts[2],
						Topic:   parts[1],
					}
					enc, _ := msg.Encode()
					p.Broadcast(enc)
				}
			case line == "/exit":
				fmt.Printf("Goodbye %s…\n", *username)
				os.Exit(0)
			}
		}
	}()

	// Block forever so container never exits
	select {}

}
