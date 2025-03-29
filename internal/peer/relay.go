package peer

import (
	"fmt"
	"net"
)

// forward sends a message to all peers except the origin.
func forward(origin net.Conn, peers []net.Conn, msg string) {
	for _, peer := range peers {
		if peer != origin {
			if _, err := fmt.Fprintln(peer, msg); err != nil {
				fmt.Println("Error forwarding message:", err)
			}
		}
	}
}
