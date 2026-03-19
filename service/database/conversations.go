package database

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
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
		if photoURL.Valid {
			s.PhotoURL = photoURL.String
		}
		if name.Valid {
			s.Name = name.String
		}

		// Impostiamo un timestamp fittizio finché non implementi la JOIN per i messaggi
		s.LastActivity = time.Now()
		summaries = append(summaries, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}

func (db *appdb) GetConversation(conversationID string, requestingUserID string) (Conversation, error) {
	var conv Conversation
	conv.ID = conversationID

	// 1. Dettagli base della conversazione (gestiamo i NullString per campi opzionali)
	var name, photo sql.NullString
	err := db.c.QueryRow("SELECT is_group, name, photo_url FROM conversations WHERE id = ?", conversationID).
		Scan(&conv.IsGroup, &name, &photo)
	if err != nil {
		return conv, err // Restituisce errore se la conversazione non esiste
	}
	if name.Valid {
		conv.Name = name.String
	}
	if photo.Valid {
		conv.PhotoURL = photo.String
	}

	// 2. CONTROLLO DI SICUREZZA: l'utente fa parte della conversazione?
	var isMember bool
	err = db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id = ? AND user_id = ?)", conversationID, requestingUserID).Scan(&isMember)
	if err != nil {
		return conv, err
	}
	if !isMember {
		return conv, errors.New("forbidden: user is not a member of this conversation")
	}

	// 3. ESTRAZIONE MEMBRI (JOIN con users)
	membersQuery := `
		SELECT u.id, u.username, u.photo_url
		FROM users u
		INNER JOIN conversation_members cm ON u.id = cm.user_id
		WHERE cm.conversation_id = ?
	`
	memberRows, err := db.c.Query(membersQuery, conversationID)
	if err != nil {
		return conv, err
	}
	defer memberRows.Close()

	conv.Members = []User{} // Inizializziamo a [] vuoto per evitare "null" nel JSON
	for memberRows.Next() {
		var u User
		var uID int64
		var uPhoto sql.NullString
		if err := memberRows.Scan(&uID, &u.Username, &uPhoto); err != nil {
			return conv, err
		}
		u.ID = strconv.FormatInt(uID, 10)
		if uPhoto.Valid {
			u.PhotoURL = uPhoto.String
		}
		conv.Members = append(conv.Members, u)
	}
	if err := memberRows.Err(); err != nil {
		return conv, err
	}

	// 4. ESTRAZIONE MESSAGGI (JOIN con users per ottenere il nome del mittente)
	messagesQuery := `
		SELECT m.id, m.content, m.message_type, m.timestamp, m.sender_id, u.username, m.photo_url, m.reply_to
		FROM messages m
		INNER JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = ?
		ORDER BY m.timestamp ASC
	`
	msgRows, err := db.c.Query(messagesQuery, conversationID)
	if err != nil {
		return conv, err
	}
	defer msgRows.Close()

	conv.Messages = []Message{}
	for msgRows.Next() {
		var m Message
		var mID, sID int64
		var content, msgPhoto sql.NullString
		var replyTo sql.NullInt64

		if err := msgRows.Scan(&mID, &content, &m.MessageType, &m.Timestamp, &sID, &m.SenderName, &msgPhoto, &replyTo); err != nil {
			return conv, err
		}

		m.ID = strconv.FormatInt(mID, 10)
		m.SenderID = strconv.FormatInt(sID, 10)
		if content.Valid {
			m.Content = content.String
		}
		if msgPhoto.Valid {
			m.PhotoURL = msgPhoto.String
		}
		if replyTo.Valid {
			m.ReplyTo = strconv.FormatInt(replyTo.Int64, 10)
		}

		conv.Messages = append(conv.Messages, m)
	}
	if err := msgRows.Err(); err != nil {
		return conv, err
	}

	return conv, nil
}
