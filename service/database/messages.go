package database

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

type Message struct {
	ID          string     `json:"id"`
	Content     string     `json:"content,omitempty"`
	MessageType string     `json:"messageType"`
	Timestamp   time.Time  `json:"timestamp"`
	SenderID    string     `json:"senderId"`
	SenderName  string     `json:"senderName"`
	Delivered   bool       `json:"delivered"`
	Read        bool       `json:"read"`
	ReplyTo     string     `json:"replyTo,omitempty"`
	Reactions   []Reaction `json:"reactions,omitempty"`
	PhotoURL    string     `json:"photoUrl,omitempty"`
}

func (db *appdb) SendMessage(conversationID string, senderID string, messageType string, content string, photoURL string, replyTo string) (Message, error) {
	// Utilizziamo NULL per replyTo se la stringa è vuota
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

	idInt, _ := res.LastInsertId()
	msgID := strconv.FormatInt(idInt, 10)

	return Message{
		ID:          msgID,
		Content:     content,
		MessageType: messageType,
		SenderID:    senderID,
		Timestamp:   time.Now(),
	}, nil
}

func (db *appdb) ForwardMessage(messageID string, targetConversationID string, senderID string) (Message, error) {
	var content, msgType string
	var photoURL sql.NullString

	// 1. Recuperiamo il contenuto e il tipo del messaggio originale
	err := db.c.QueryRow("SELECT content, message_type, photo_url FROM messages WHERE id = ?", messageID).
		Scan(&content, &msgType, &photoURL)
	if err != nil {
		return Message{}, err
	}

	// 2. Usiamo SendMessage per inserire il nuovo messaggio
	// Passiamo una stringa vuota per replyTo poiché è un inoltro
	return db.SendMessage(targetConversationID, senderID, msgType, content, photoURL.String, "")
}
func (db *appdb) DeleteMessage(messageID string, requestingUserID string) error {
	res, err := db.c.Exec("DELETE FROM messages WHERE id = ? AND sender_id = ?", messageID, requestingUserID)
	if err != nil {
		return err // Errore vero e proprio del database (es. sintassi o connessione persa)
	}

	// Controlliamo se abbiamo effettivamente cancellato qualcosa
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		// Se affected è 0, il messaggio non esiste o l'utente non è il mittente!
		return errors.New("message not found or forbidden")
	}

	return nil
}
