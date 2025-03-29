package db

import (
	"database/sql"
	"fmt"

	"github.com/2004joshua/nodechat/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB ...
func InitDB(filepath string) error {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		return err
	}

	sqlStmt := `
    CREATE TABLE IF NOT EXISTS messages (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        type TEXT,
        sender TEXT,
        content TEXT,
        timestamp INTEGER,
        topic TEXT
    );
    `
	_, err = DB.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}

// SaveMessage persists a Message to the SQLite database.
func SaveMessage(msg *model.Message) error {
	stmt, err := DB.Prepare("INSERT INTO messages(type, sender, content, timestamp, topic) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(msg.Type, msg.Sender, msg.Content, msg.Timestamp, msg.Topic)
	if err != nil {
		return err
	}
	return nil
}
