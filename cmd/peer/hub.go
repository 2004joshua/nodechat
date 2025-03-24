package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Maintains a set of active connections and broadcasts messages
type Hub struct {
	clients    map[*websocket.Conn]bool // active connections
	broadcast  chan []byte              // channel for messages for all clients
	register   chan *websocket.Conn     // channel for registering new clients
	unregister chan *websocket.Conn     // channel for unregistering clients
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			// new client is connected
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()
		case conn := <-h.unregister:
			// client is disconnected
			h.mu.Lock()
			delete(h.clients, conn)
			conn.Close()
			h.mu.Unlock()
		case msg := <-h.broadcast:
			// broadcast message to all clients
			h.mu.Lock()
			for conn := range h.clients {
				conn.WriteMessage(websocket.TextMessage, msg)
			}
			h.mu.Unlock()
		}
	}
}

// configures websocket from http
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// HandleWS handles websocket connections
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil) // websocket handshake
	if err != nil {
		return
	}
	h.register <- conn

	// listen for messages
	go func() {
		defer func() { h.unregister <- conn }()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			h.broadcast <- msg
		}
	}()
}
