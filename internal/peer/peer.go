package peer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/2004joshua/nodechat/internal/storage"
)

// Peer represents a local node in the P2P network.
type Peer struct {
	Addr  string      // Listening address
	peers []net.Conn  // Active peer connections
	mu    sync.Mutex  // Protects the peers slice
	DB    *storage.DB // Local database for message storage
	Name  string      // This peer's name
}

// New creates a new Peer instance.
func New(addr, name string, db *storage.DB) *Peer {
	return &Peer{
		Addr:  addr,
		peers: make([]net.Conn, 0),
		DB:    db,
		Name:  name,
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
			// Send any offline (undelivered) messages to this new peer.
			p.sendOfflineMessages(conn)
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
	// Send any offline (undelivered) messages to this peer.
	p.sendOfflineMessages(conn)
	return nil
}

func (p *Peer) addPeer(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = append(p.peers, conn)
}

func (p *Peer) removePeer(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, c := range p.peers {
		if c == conn {
			p.peers = append(p.peers[:i], p.peers[i+1:]...)
			break
		}
	}
}

// handleConn processes incoming messages from a peer.
func (p *Peer) handleConn(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		var msg Message
		err := json.Unmarshal([]byte(line), &msg)
		if err != nil {
			fmt.Println("Error decoding JSON message:", err)
			continue
		}
		// Display the message.
		fmt.Printf("[%s] %s: %s\n", conn.RemoteAddr(), msg.Sender, msg.Content)
		// Save the received message as delivered.
		if p.DB != nil {
			err := p.DB.SaveMessage(msg.Sender, msg.Content, msg.Timestamp, true)
			if err != nil {
				fmt.Println("Error saving received message:", err)
			}
		}
		// Forward the message to all other connected peers.
		forward(conn, p.peers, line)
	}
	p.removePeer(conn)
	conn.Close()
}

// Broadcast sends a message to all connected peers.
func (p *Peer) Broadcast(msgText string) {
	msg := Message{
		Type:      "chat",
		Sender:    p.Name,
		Timestamp: time.Now(),
		Content:   msgText,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error encoding message:", err)
		return
	}
	// Save the broadcast message locally.
	if p.DB != nil {
		err = p.DB.SaveMessage(p.Name, msg.Content, msg.Timestamp, true)
		if err != nil {
			fmt.Println("Error saving broadcast message:", err)
		}
	}
	// Forward the message to all connected peers.
	forward(nil, p.peers, string(msgBytes))
}

// sendOfflineMessages sends stored undelivered messages to a given connection.
func (p *Peer) sendOfflineMessages(conn net.Conn) {
	if p.DB == nil {
		return
	}
	messages, err := p.DB.GetUndeliveredMessages()
	if err != nil {
		fmt.Println("Error retrieving offline messages:", err)
		return
	}
	for _, m := range messages {
		// Reconstruct the Message struct.
		msg := Message{
			Type:      "chat",
			Sender:    m.Sender,
			Timestamp: m.Timestamp,
			Content:   m.Content,
		}
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Error encoding offline message:", err)
			continue
		}
		// Send the message to the peer.
		fmt.Fprintln(conn, string(msgBytes))
		// Mark the message as delivered.
		p.DB.MarkMessageDelivered(m.ID)
	}
}
