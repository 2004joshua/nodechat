package peer

import (
	"fmt"
	"net"
)

// forward sends msg to all peers except the origin
func forward(origin net.Conn, peers []net.Conn, msg string) {
	for _, p := range peers {
		if p != origin {
			fmt.Fprintln(p, msg)
		}
	}
}
