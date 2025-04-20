package model

import (
	"encoding/json"
	"time"
)

type Message struct {
	Type      string `json:"type"`               // "chat", "file", etc.
	Content   string `json:"content,omitempty"`  // only for chat
	FileURL   string `json:"fileUrl,omitempty"`  // only for files
	FileName  string `json:"fileName,omitempty"` // only for files
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
	var msg Message
	err := json.Unmarshal([]byte(data), &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
