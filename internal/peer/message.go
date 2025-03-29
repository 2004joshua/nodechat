package peer

import "time"

// Message defines our messaging protocol.
type Message struct {
	Type      string    `json:"type"`
	Sender    string    `json:"sender"`
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
}
