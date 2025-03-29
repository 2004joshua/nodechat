package peer

import (
	"bufio"
	"fmt"
	"net"

	"github.com/2004joshua/nodechat/internal/db"
	"github.com/2004joshua/nodechat/internal/model"
)

// Peer ...
type Peer struct {
	Addr  string
	peers []net.Conn
}

func New(addr string) *Peer {
	return &Peer{
		Addr:  addr,
		peers: make([]net.Conn, 0),
	}
}

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

func (p *Peer) Broadcast(msg string) {
	for _, conn := range p.peers {
		fmt.Fprintln(conn, msg)
	}
}

func (p *Peer) handleConn(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		data := scanner.Text()

		// Use model.DecodeMessage instead of peer.DecodeMessage
		msg, err := model.DecodeMessage(data)
		if err != nil {
			// Fallback: treat as plain text if JSON decoding fails.
			fmt.Printf("[%s] %s\n", conn.RemoteAddr(), data)
			continue
		}

		switch msg.Type {
		case "chat":
			fmt.Printf("[%s] %s: %s\n", conn.RemoteAddr(), msg.Sender, msg.Content)
		case "notification":
			fmt.Printf("[Notification] %s: %s\n", msg.Sender, msg.Content)
		case "command":
			fmt.Printf("[Command] %s: %s\n", msg.Sender, msg.Content)
		default:
			fmt.Printf("[%s] Unknown message type: %s\n", conn.RemoteAddr(), msg.Content)
		}

		// Save the message
		if err := db.SaveMessage(msg); err != nil {
			fmt.Println("Error saving message to database:", err)
		}

		// Forward the raw JSON message to other peers
		forward(conn, p.peers, data)
	}
	p.removePeer(conn)
	conn.Close()
}
