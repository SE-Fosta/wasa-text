package database

import (
	"database/sql"
	"time"
	"strconv"
)

type MessagePreview struct {
	Content     string `json:"content,omitempty"`
	MessageType string `json:"messageType"`
	SenderName  string `json:"senderName,omitempty"`
}

type ConversationSummary struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	PhotoURL     string         `json:"photoUrl,omitempty"`
	LastMessage  MessagePreview `json:"lastMessage,omitempty"`
	LastActivity time.Time      `json:"lastActivity"`
	IsGroup      bool           `json:"isGroup"`
	UnreadCount  int            `json:"unreadCount"`
}

type Conversation struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	PhotoURL string    `json:"photoUrl,omitempty"`
	IsGroup  bool      `json:"isGroup"`
	Members  []User    `json:"members"`
	Messages []Message `json:"messages"`
}

func (db *appdb) GetMyConversations(userID string) ([]ConversationSummary, error) {
	// Query base: recupera le conversazioni di cui l'utente fa parte.
	// Nota: qui andrà aggiunta la JOIN per recuperare l'ultimo messaggio (LastMessage) per mostrare l'anteprima.
	query := `
		SELECT c.id, c.name, c.photo_url, c.is_group
		FROM conversations c
		INNER JOIN conversation_members cm ON c.id = cm.conversation_id
		WHERE cm.user_id = ?
	`
	rows, err := db.c.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []ConversationSummary
	for rows.Next() {
		var s ConversationSummary
		var photoURL, name sql.NullString
		var idInt int64
		if err := rows.Scan(&idInt, &name, &photoURL, &s.IsGroup); err != nil {
			return nil, err
		}
		// Convertiamo l'ID intero in stringa
		s.ID = strconv.FormatInt(idInt, 10)
		if photoURL.Valid { s.PhotoURL = photoURL.String }
		if name.Valid { s.Name = name.String }
		
		// Impostiamo un timestamp fittizio finché non implementi la JOIN per i messaggi
		s.LastActivity = time.Now() 
		summaries = append(summaries, s)
	}
	return summaries, nil
}

func (db *appdb) GetConversation(conversationID string, requestingUserID string) (Conversation, error) {
	var conv Conversation
	conv.ID = conversationID
	
	err := db.c.QueryRow("SELECT is_group, name, photo_url FROM conversations WHERE id = ?", conversationID).
		Scan(&conv.IsGroup, &conv.Name, &conv.PhotoURL)
	if err != nil {
		return conv, err
	}

	// TODO: Completare con l'estrazione dei membri (JOIN users) e dei messaggi (SELECT su messages)
	return conv, nil
}