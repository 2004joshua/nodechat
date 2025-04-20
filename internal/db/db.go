package db

import (
	"database/sql"
	"fmt"

	"github.com/2004joshua/nodechat/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB opens (or creates) the SQLite database and sets up the messages table.
func InitDB(filepath string) error {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		return err
	}

	// Create messages and subscriptions tables, with file support
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		type TEXT,
		sender TEXT,
		content TEXT,
		timestamp INTEGER,
		topic TEXT,
		file_url TEXT,
		file_name TEXT
	);

	CREATE TABLE IF NOT EXISTS subscriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		topic TEXT
	);
	`
	_, err = DB.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	return nil
}

// SaveMessage persists a Message to the SQLite database.
func SaveMessage(msg *model.Message) error {
	stmt, err := DB.Prepare(
		`INSERT INTO messages(type, sender, content, timestamp, topic, file_url, file_name)
		 VALUES(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		msg.Type,
		msg.Sender,
		msg.Content,
		msg.Timestamp,
		msg.Topic,
		msg.FileURL,
		msg.FileName,
	)
	return err
}

// GetMessages retrieves all messages ordered by timestamp.
func GetMessages() ([]model.Message, error) {
	rows, err := DB.Query(
		`SELECT type, sender, content, timestamp, topic, file_url, file_name
		  FROM messages
		 ORDER BY timestamp ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var m model.Message
		err := rows.Scan(
			&m.Type,
			&m.Sender,
			&m.Content,
			&m.Timestamp,
			&m.Topic,
			&m.FileURL,
			&m.FileName,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

// SaveSubscription stores a new subscription.
func SaveSubscription(username, topic string) error {
	stmt, err := DB.Prepare("INSERT INTO subscriptions(username, topic) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(username, topic)
	return err
}

// RemoveSubscription deletes a topic subscription.
func RemoveSubscription(username, topic string) error {
	stmt, err := DB.Prepare("DELETE FROM subscriptions WHERE username = ? AND topic = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(username, topic)
	return err
}

// GetSubscriptions returns a list of topics for a user.
func GetSubscriptions(username string) ([]string, error) {
	rows, err := DB.Query("SELECT topic FROM subscriptions WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var topic string
		if err := rows.Scan(&topic); err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}
	return topics, nil
}
