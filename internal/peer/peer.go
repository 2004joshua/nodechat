package peer

import (
	"bufio"
	"fmt"
	"net"

	"github.com/2004joshua/nodechat/internal/db"
	"github.com/2004joshua/nodechat/internal/model"
)

// Peer represents a node in the P2P network.
type Peer struct {
	Addr  string
	peers []net.Conn
}

// New creates a new Peer listening on the given address.
func New(addr string) *Peer {
	return &Peer{
		Addr:  addr,
		peers: make([]net.Conn, 0),
	}
}

// Listen starts accepting incoming connections.
func (p *Peer) Listen() error {
	ln, err := net.Listen("tcp", p.Addr)
	if err != nil {
		return err
	}
	fmt.Println("Listening on", p.Addr)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}
			p.addPeer(conn)
			fmt.Println("Incoming connection from", conn.RemoteAddr())
			go p.handleConn(conn)
		}
	}()
	return nil
}

// Connect dials out to another peer.
func (p *Peer) Connect(remote string) error {
	conn, err := net.Dial("tcp", remote)
	if err != nil {
		return err
	}
	p.addPeer(conn)
	fmt.Println("Connected to", remote)
	go p.handleConn(conn)
	return nil
}

func (p *Peer) addPeer(conn net.Conn) {
	p.peers = append(p.peers, conn)
}

func (p *Peer) removePeer(conn net.Conn) {
	for i, c := range p.peers {
		if c == conn {
			p.peers = append(p.peers[:i], p.peers[i+1:]...)
			break
		}
	}
}

// Broadcast sends a message (already JSON-encoded) to all connected peers.
func (p *Peer) Broadcast(msg string) {
	for _, conn := range p.peers {
		fmt.Fprintln(conn, msg)
	}
}

// handleConn processes incoming messages from a single connection.
func (p *Peer) handleConn(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		data := scanner.Text()

		// Attempt to decode data as a JSON message.
		msg, err := model.DecodeMessage(data)
		if err != nil {
			// Fallback: treat as plain text if JSON decoding fails.
			fmt.Printf("[%s] %s\n", conn.RemoteAddr(), data)
			continue
		}

		// Process the message based on its type.
		p.processMessage(msg, conn.RemoteAddr())

		// Save the message to the database.
		if err := db.SaveMessage(msg); err != nil {
			fmt.Println("Error saving message to database:", err)
		}

		// Forward the raw JSON message to other connected peers.
		forward(conn, p.peers, data)
	}
	p.removePeer(conn)
	conn.Close()
}

// processMessage handles incoming messages based on their type.
func (p *Peer) processMessage(msg *model.Message, senderAddr net.Addr) {
	switch msg.Type {
	case "chat":
		fmt.Printf("[%s] %s: %s\n", senderAddr, msg.Sender, msg.Content)
	case "notification":
		fmt.Printf("[Notification] %s: %s\n", msg.Sender, msg.Content)
	case "command":
		fmt.Printf("[Command] %s: %s\n", msg.Sender, msg.Content)
		// Example automation: if the command is "ping", reply with "pong"
		if msg.Content == "ping" {
			response := &model.Message{
				Type:    "command",
				Sender:  "self", // Adjust as necessary for your system
				Content: "pong",
			}
			encoded, err := response.Encode()
			if err != nil {
				fmt.Println("Error encoding command response:", err)
			} else {
				// Broadcast the response (or you could send it directly to the sender)
				p.Broadcast(encoded)
			}
		}
	default:
		fmt.Printf("[%s] Unknown message type: %s\n", senderAddr, msg.Content)
	}
}
