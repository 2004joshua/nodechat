package peer

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Peer struct {
	Addr  string
	peers []net.Conn
	mu    sync.Mutex
}

// New creates a new Peer listening on given host:port
func New(addr string) *Peer {
	return &Peer{Addr: addr, peers: make([]net.Conn, 0)}
}

// Listen starts accepting incoming connections
func (p *Peer) Listen() error {
	ln, err := net.Listen("tcp", p.Addr)
	if err != nil {
		return err
	}
	fmt.Println("Listening on", p.Addr)
	go func() {
		for {
			conn, _ := ln.Accept()
			p.addPeer(conn)
			fmt.Println("Incoming from", conn.RemoteAddr())
			go p.handleConn(conn)
		}
	}()
	return nil
}

// Connect dials out to another peer
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

func (p *Peer) handleConn(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		fmt.Printf("[%s] %s\n", conn.RemoteAddr(), msg)
		forward(conn, p.peers, msg)
	}
	p.removePeer(conn)
	conn.Close()
}

// Broadcast sends a message to all connected peers
func (p *Peer) Broadcast(msg string) {
	forward(nil, p.peers, msg)
}
