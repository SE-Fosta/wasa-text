package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AppDatabase is the high level interface for the DB.
type AppDatabase interface {
	Ping() error

	// -- Utenti --
	DoLogin(username string) (string, error)
	SetMyUserName(userID string, newName string) error
	SetMyPhoto(userID string, photoURL string) error

	// -- Conversazioni --
	GetMyConversations(userID string) ([]ConversationSummary, error)
	GetConversation(conversationID string, requestingUserID string) (Conversation, error)

	// -- Messaggi --
	SendMessage(conversationID string, senderID string, messageType string, content string, photoURL string, replyTo string) (Message, error)
	ForwardMessage(messageID string, targetConversationID string, senderID string) (Message, error)
	DeleteMessage(messageID string, requestingUserID string) error
	CommentMessage(messageID string, userID string, emoji string) error
	UncommentMessage(messageID string, userID string) error

	// -- Gruppi --
	AddToGroup(groupID string, userIDToAdd string) error
	LeaveGroup(groupID string, userID string) error
	SetGroupName(groupID string, newName string) error
	SetGroupPhoto(groupID string, photoURL string) error
}

type appdb struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building an AppDatabase")
	}

	// Abilitiamo le Foreign Keys
	_, err := db.Exec(`PRAGMA foreign_keys = ON;`)
	if err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	var tableName string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		// Creazione tabelle con AUTOINCREMENT per gli ID
		sqlStmt := `
		CREATE TABLE users (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			photo_url TEXT
		);

		CREATE TABLE conversations (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			is_group BOOLEAN NOT NULL DEFAULT 0,
			name TEXT,
			photo_url TEXT
		);

		CREATE TABLE conversation_members (
			conversation_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			PRIMARY KEY (conversation_id, user_id),
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE messages (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			conversation_id INTEGER NOT NULL,
			sender_id INTEGER NOT NULL,
			content TEXT,
			message_type TEXT NOT NULL,
			photo_url TEXT,
			reply_to INTEGER,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY(sender_id) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE message_status (
			message_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			delivered BOOLEAN DEFAULT 0,
			read BOOLEAN DEFAULT 0,
			PRIMARY KEY (message_id, user_id),
			FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE reactions (
			message_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			emoji TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (message_id, user_id),
			FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error checking database structure: %w", err)
	}

	return &appdb{c: db}, nil
}

func (db *appdb) Ping() error {
	return db.c.Ping()
}
