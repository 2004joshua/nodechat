package peer

import (
	"fmt"
	"net"
)

// forward sends msg to all peers except origin
func forward(origin net.Conn, peers []net.Conn, msg string) {
	for _, peer := range peers {
		if peer != origin {
			fmt.Fprintln(peer, msg)
		}
	}
}
