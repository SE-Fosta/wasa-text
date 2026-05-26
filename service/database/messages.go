package database

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"
)

type Message struct {
	ID             string     `json:"id"`
	Content        string     `json:"content,omitempty"`
	MessageType    string     `json:"messageType"`
	Timestamp      time.Time  `json:"timestamp"`
	SenderID       string     `json:"senderId"`
	SenderName     string     `json:"senderName"`
	SenderPhotoUrl string     `json:"senderPhotoUrl"`
	Delivered      bool       `json:"delivered"`
	Read           bool       `json:"read"`
	ReplyTo        string     `json:"replyTo,omitempty"`
	Reactions      []Reaction `json:"reactions,omitempty"`
	PhotoURL       string     `json:"photoUrl,omitempty"`
}

type Reaction struct {
	UserID string `json:"userId"`
	Emoji  string `json:"emoji"`
}

func (db *appdb) SendMessage(conversationID string, senderID string, messageType string, content string, photoURL string, replyTo string) (Message, error) {
	var replyToVal interface{} = replyTo
	if replyTo == "" {
		replyToVal = nil
	}

	res, err := db.c.Exec(`
        INSERT INTO messages (conversation_id, sender_id, content, message_type, photo_url, reply_to) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		conversationID, senderID, content, messageType, photoURL, replyToVal)

	if err != nil {
		return Message{}, err
	}

	lastID, _ := res.LastInsertId()
	msgIDStr := strconv.FormatInt(lastID, 10)

	_, err = db.c.Exec(`
        INSERT INTO message_status (message_id, user_id, `+"`delivered`"+`, `+"`read`"+`)
        SELECT ?, user_id, 0, 0 
        FROM conversation_members 
        WHERE conversation_id = ? AND user_id != ?`,
		lastID, conversationID, senderID)

	if err != nil {
		log.Printf("Errore creazione message_status: %v\n", err)
	}

	return Message{
		ID:          msgIDStr,
		Content:     content,
		MessageType: messageType,
		SenderID:    senderID,
		PhotoURL:    photoURL,
		Timestamp:   time.Now(),
		Delivered:   false,
		Read:        false,
	}, nil
}
func (db *appdb) ForwardMessage(messageID string, targetConversationID string, senderID string) (Message, error) {
	var msgType string
	var content, photoURL sql.NullString

	err := db.c.QueryRow("SELECT content, message_type, photo_url FROM messages WHERE id = ?", messageID).
		Scan(&content, &msgType, &photoURL)
	if err != nil {
		return Message{}, err
	}

	safeContent := ""
	if content.Valid {
		safeContent = content.String
	}

	safePhotoURL := ""
	if photoURL.Valid {
		safePhotoURL = photoURL.String
	}

	return db.SendMessage(targetConversationID, senderID, msgType, safeContent, safePhotoURL, "")
}

func (db *appdb) DeleteMessage(messageID string, requestingUserID string) error {
	res, err := db.c.Exec("DELETE FROM messages WHERE id = ? AND sender_id = ?", messageID, requestingUserID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("message not found or forbidden")
	}

	return nil
}

func (db *appdb) GetMessages(conversationID string) ([]Message, error) {
	query := `
        SELECT 
            m.id, m.content, m.message_type, m.timestamp, m.sender_id, 
            u.username as sender_name, 
            u.photo_url as sender_photo_url,
            m.reply_to, 
            m.photo_url,
            COALESCE(MIN(ms.delivered), 0), 
            COALESCE(MIN(ms.read), 0)
        FROM messages m
        INNER JOIN users u ON m.sender_id = u.id
        LEFT JOIN message_status ms ON m.id = ms.message_id AND ms.user_id != m.sender_id
        WHERE m.conversation_id = ?
        GROUP BY m.id, m.content, m.message_type, m.timestamp, m.sender_id, u.username, u.photo_url, m.reply_to, m.photo_url
        ORDER BY m.timestamp ASC
    `
	rows, err := db.c.Query(query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		var content, replyTo, photoURL, senderPhotoURL sql.NullString

		err := rows.Scan(
			&m.ID,
			&content,
			&m.MessageType,
			&m.Timestamp,
			&m.SenderID,
			&m.SenderName,
			&senderPhotoURL,
			&replyTo,
			&photoURL,
			&m.Delivered,
			&m.Read,
		)
		if err != nil {
			return nil, err
		}

		if content.Valid {
			m.Content = content.String
		}
		if replyTo.Valid {
			m.ReplyTo = replyTo.String
		}
		if photoURL.Valid {
			m.PhotoURL = photoURL.String
		}

		if senderPhotoURL.Valid {
			m.SenderPhotoUrl = senderPhotoURL.String
		}

		m.Reactions = make([]Reaction, 0)
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(messages) > 0 {
		reactionQuery := `
            SELECT r.message_id, r.user_id, r.emoji
            FROM reactions r
            INNER JOIN messages m ON r.message_id = m.id
            WHERE m.conversation_id = ?
        `
		rRows, err := db.c.Query(reactionQuery, conversationID)
		if err != nil {
			return nil, err
		}
		defer rRows.Close()

		reactionsMap := make(map[string][]Reaction)

		for rRows.Next() {
			var msgID, userID, emoji string
			if err := rRows.Scan(&msgID, &userID, &emoji); err != nil {
				return nil, err
			}

			reactionsMap[msgID] = append(reactionsMap[msgID], Reaction{
				UserID: userID,
				Emoji:  emoji,
			})
		}

		if err = rRows.Err(); err != nil {
			return nil, err
		}

		for i, msg := range messages {
			if rx, found := reactionsMap[msg.ID]; found {
				messages[i].Reactions = rx
			}
		}
	}

	return messages, nil
}
func (db *appdb) MarkAsRead(conversationID string, userID string) error {
	query := `
        UPDATE message_status 
        SET read = 1, delivered = 1 
        WHERE user_id = ? 
        AND message_id IN (
            SELECT id FROM messages WHERE conversation_id = ? AND sender_id != ?
        )`

	_, err := db.c.Exec(query, userID, conversationID, userID)
	if err != nil {
		log.Printf("--- ERRORE SQL MarkAsRead: %v ---\n", err)
		return err
	}
	return nil
}

func (db *appdb) ReactMessage(messageID string, userID string, emoji string) error {
	query := `
		INSERT INTO reactions (message_id, user_id, emoji) 
		VALUES (?, ?, ?)
		ON CONFLICT(message_id, user_id) 
		DO UPDATE SET 
			emoji = excluded.emoji,
			timestamp = CURRENT_TIMESTAMP;
	`
	_, err := db.c.Exec(query, messageID, userID, emoji)
	return err
}

func (db *appdb) UnreactMessage(messageID string, userID string) error {
	query := `
		DELETE FROM reactions 
		WHERE message_id = ? AND user_id = ?
	`
	_, err := db.c.Exec(query, messageID, userID)
	return err
}

func (db *appdb) CommentMessage(originalMessageID string, senderID string, content string) (Message, error) {
	var conversationID string
	err := db.c.QueryRow("SELECT conversation_id FROM messages WHERE id = ?", originalMessageID).Scan(&conversationID)
	if err != nil {
		return Message{}, err
	}

	return db.SendMessage(conversationID, senderID, "text", content, "", originalMessageID)
}

func (db *appdb) UncommentMessage(commentID string, requestingUserID string) error {
	return db.DeleteMessage(commentID, requestingUserID)
}
