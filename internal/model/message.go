package model

import (
	"encoding/json"
	"time"
)

type Message struct {
	Type      string `json:"type"` // e.g. "chat", "notification", "command"
	Content   string `json:"content"`
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
	Topic     string `json:"topic,omitempty"`
}

// Encode converts a Message to a JSON string.
func (m *Message) Encode() (string, error) {
	m.Timestamp = time.Now().Unix()
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DecodeMessage converts a JSON string back into a Message.
func DecodeMessage(data string) (*Message, error) {
	var m Message
	err := json.Unmarshal([]byte(data), &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
