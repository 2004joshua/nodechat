package storage

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQL connection.
type DB struct {
	Conn *sql.DB
}

// StoredMessage represents a message stored in the database.
type StoredMessage struct {
	ID        int
	Sender    string
	Timestamp time.Time
	Content   string
	Delivered bool
}

// NewDB initializes the database connection and creates the messages table if it doesn't exist.
func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	createTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender TEXT,
		timestamp DATETIME,
		content TEXT,
		delivered BOOLEAN
	);
	`
	_, err = conn.Exec(createTable)
	if err != nil {
		return nil, err
	}
	return &DB{Conn: conn}, nil
}

// SaveMessage stores a message in the database.
func (db *DB) SaveMessage(sender, content string, timestamp time.Time, delivered bool) error {
	stmt, err := db.Conn.Prepare("INSERT INTO messages(sender, timestamp, content, delivered) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(sender, timestamp.Format("2006-01-02 15:04:05"), content, delivered)
	return err
}

// GetUndeliveredMessages retrieves messages that haven't been marked as delivered.
func (db *DB) GetUndeliveredMessages() ([]StoredMessage, error) {
	rows, err := db.Conn.Query("SELECT id, sender, timestamp, content, delivered FROM messages WHERE delivered = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []StoredMessage
	for rows.Next() {
		var m StoredMessage
		var ts string
		if err := rows.Scan(&m.ID, &m.Sender, &ts, &m.Content, &m.Delivered); err != nil {
			return nil, err
		}
		m.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		messages = append(messages, m)
	}
	return messages, nil
}

// MarkMessageDelivered flags a message as delivered.
func (db *DB) MarkMessageDelivered(id int) error {
	_, err := db.Conn.Exec("UPDATE messages SET delivered = 1 WHERE id = ?", id)
	return err
}
