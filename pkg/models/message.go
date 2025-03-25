package models

import "time"

// structure of message
type Message struct {
	Sender    string    `json:"sender"`
	Recipient string    `json:"recipient"`
	Body      string    `json:"body"`
	Timestamp time.Time `json:"timestamp"`
}
