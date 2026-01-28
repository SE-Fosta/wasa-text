package database

import (
	"database/sql"
	"errors"
	"time"
	"fmt"
)

// Definiamo le struct di dominio che useremo per scambiare dati con l'API
// Nota: Alcuni campi json corrispondono allo schema API

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	PhotoURL string `json:"photoUrl,omitempty"`
}

type ConversationSummary struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	PhotoURL     string    `json:"photoUrl"`
	LastMessage  string    `json:"lastMessage"` // Anteprima del contenuto
	LastActivity time.Time `json:"lastActivity"`
	IsGroup      bool      `json:"isGroup"`
	UnreadCount  int       `json:"unreadCount"` // Opzionale, mettiamo 0 per semplicità se non gestito
}

type Message struct {
	ID          string     `json:"id"`
	ConversationID string  `json:"-"` // Interuso interno
	Content     string     `json:"content"`
	MessageType string     `json:"messageType"`
	Timestamp   time.Time  `json:"timestamp"`
	SenderID    string     `json:"senderId"`
	SenderName  string     `json:"senderName"`
	ReplyTo     *string    `json:"replyTo,omitempty"`
	PhotoURL    string     `json:"photoUrl,omitempty"`
	Reactions   []Reaction `json:"reactions"`
	Delivered   bool       `json:"delivered"`
	Read        bool       `json:"read"`
}

type Reaction struct {
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Emoji     string    `json:"emoji"`
	Timestamp time.Time `json:"timestamp"` // Aggiunta se serve, o usa time.Now()
}

// DoLogin (già vista) - Gestisce il login semplificato
func DoLogin(username string) (string, error) {
	var id string
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Crea nuovo utente se non esiste
			newID := fmt.Sprintf("%d", time.Now().UnixNano()) // Semplice ID generator o usa UUID
			_, err := db.Exec("INSERT INTO users (id, username) VALUES (?, ?)", newID, username)
			if err != nil {
				return "", err
			}
			return newID, nil
		}
		return "", err
	}
	return id, nil
}

// SetMyUserName - Aggiorna il nome utente
func SetMyUserName(userID string, newName string) error {
	// Verifica unicità
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? AND id != ?", newName, userID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("username already taken")
	}
	_, err = db.Exec("UPDATE users SET username = ? WHERE id = ?", newName, userID)
	return err
}

// SetMyPhoto - Aggiorna la foto profilo (salviamo l'URL o il path)
func SetMyPhoto(userID string, photoURL string) error {
	_, err := db.Exec("UPDATE users SET photo_url = ? WHERE id = ?", photoURL, userID)
	return err
}

// GetMyConversations - Ritorna la lista delle conversazioni per la dashboard
// Questa è la query più complessa: deve recuperare le conversazioni dell'utente,
// trovare l'ultimo messaggio per ciascuna e ordinare per data decrescente.
func GetMyConversations(userID string) ([]ConversationSummary, error) {
	// Query: Seleziona conversazioni a cui l'utente partecipa
	// Nota: Per semplicità, qui facciamo una query base e poi popoliamo l'ultimo messaggio
	// In produzione si userebbero Window Functions o Join complessi.

	query := `
	SELECT c.id, c.name, c.is_group, c.photo_url
	FROM conversations c
	JOIN participants p ON c.id = p.conversation_id
	WHERE p.user_id = ?
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []ConversationSummary

	for rows.Next() {
		var c ConversationSummary
		var photoURL sql.NullString
		var name sql.NullString

		if err := rows.Scan(&c.ID, &name, &c.IsGroup, &photoURL); err != nil {
			return nil, err
		}
		c.PhotoURL = photoURL.String
		c.Name = name.String

		// Recuperiamo l'ultimo messaggio per questa conversazione
		// (Ottimizzazione: questa query N+1 va bene per piccoli volumi, per esame è ok)
		var msgContent, msgType, senderName string
		var msgTime time.Time
		err = db.QueryRow(`
		SELECT m.content, m.message_type, m.created_at, u.username
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = ?
		ORDER BY m.created_at DESC LIMIT 1
		`, c.ID).Scan(&msgContent, &msgType, &msgTime, &senderName)

		if err == nil {
			c.LastActivity = msgTime
			if msgType == "photo" {
				c.LastMessage = "Sent a photo"
			} else {
				c.LastMessage = msgContent
			}
		} else {
			// Nessun messaggio ancora
			c.LastActivity = time.Now() // O data creazione gruppo
			c.LastMessage = ""
		}

		conversations = append(conversations, c)
	}

	// Qui dovresti ordinare 'conversations' per LastActivity DESC in Go,
	// oppure rifinire la query SQL per farlo direttamente.

	return conversations, nil
}

// GetConversation - Ritorna tutti i messaggi di una conversazione
func GetConversation(conversationID string) ([]Message, []User, error) { // Ritorna messaggi e membri
	// 1. Recupera i messaggi
	queryMsgs := `
	SELECT m.id, m.content, m.message_type, m.created_at, m.sender_id, u.username, m.photo_url, m.reply_to
	FROM messages m
	JOIN users u ON m.sender_id = u.id
	WHERE m.conversation_id = ?
	ORDER BY m.created_at DESC
	`
	rows, err := db.Query(queryMsgs, conversationID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		var replyTo sql.NullString
		var photoURL sql.NullString
		if err := rows.Scan(&m.ID, &m.Content, &m.MessageType, &m.Timestamp, &m.SenderID, &m.SenderName, &photoURL, &replyTo); err != nil {
			return nil, nil, err
		}
		if replyTo.Valid { val := replyTo.String; m.ReplyTo = &val }
		m.PhotoURL = photoURL.String

		// Carica reazioni per il messaggio (query N+1 semplificata)
		reacs, _ := getReactionsForMessage(m.ID)
		m.Reactions = reacs

		messages = append(messages, m)
	}

	// 2. Recupera i membri (partecipanti)
	queryMembers := `
	SELECT u.id, u.username, u.photo_url
	FROM participants p
	JOIN users u ON p.user_id = u.id
	WHERE p.conversation_id = ?
	`
	rowMembers, err := db.Query(queryMembers, conversationID)
	if err != nil {
		return messages, nil, nil // Ritorna messaggi anche se fallisce membri (o gestisci errore)
	}
	defer rowMembers.Close()

	var members []User
	for rowMembers.Next() {
		var u User
		var photo sql.NullString
		if err := rowMembers.Scan(&u.ID, &u.Username, &photo); err != nil {
			continue
		}
		u.PhotoURL = photo.String
		members = append(members, u)
	}

	return messages, members, nil
}

// Helper per le reazioni
func getReactionsForMessage(messageID string) ([]Reaction, error) {
	rows, err := db.Query(`
	SELECT r.user_id, u.username, r.emoji
	FROM reactions r
	JOIN users u ON r.user_id = u.id
	WHERE r.message_id = ?`, messageID)
	if err != nil { return nil, err }
	defer rows.Close()

	var reactions []Reaction
	for rows.Next() {
		var r Reaction
		if err := rows.Scan(&r.UserID, &r.UserName, &r.Emoji); err == nil {
			reactions = append(reactions, r)
		}
	}
	return reactions, nil
}

// SendMessage - Invia un messaggio
func SendMessage(conversationID string, senderID string, content string, msgType string, photoURL string, replyTo string) (Message, error) {
	// Genera ID messaggio
	msgID := fmt.Sprintf("%d", time.Now().UnixNano())
	createdAt := time.Now()

	// Controllo esistenza conversazione (opzionale: se 1-to-1 potrebbe non esistere ancora nel DB)
	// Per semplicità assumiamo che la conversazione sia stata creata o che la creiamo qui se necessario.

	var replyToVal interface{}
	if replyTo != "" { replyToVal = replyTo } else { replyToVal = nil }

	_, err := db.Exec(`
	INSERT INTO messages (id, conversation_id, sender_id, content, message_type, photo_url, reply_to, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, msgID, conversationID, senderID, content, msgType, photoURL, replyToVal, createdAt)

	if err != nil {
		return Message{}, err
	}

	// Aggiorna timestamp ultima attività conversazione (utile per l'ordinamento)
	db.Exec("UPDATE conversations SET last_message_at = ? WHERE id = ?", createdAt, conversationID)

	// Costruisci oggetto risposta
	var senderName string
	db.QueryRow("SELECT username FROM users WHERE id = ?", senderID).Scan(&senderName)

	return Message{
		ID: msgID, ConversationID: conversationID, Content: content,
		MessageType: msgType, Timestamp: createdAt, SenderID: senderID,
		SenderName: senderName, PhotoURL: photoURL,
		Delivered: true, Read: false, // Default
	}, nil
}

// ForwardMessage - Inoltra un messaggio esistente
func ForwardMessage(originalMsgID string, targetConversationID string, newSenderID string) error {
	// Recupera contenuto originale
	var content, msgType, photoURL string
	err := db.QueryRow("SELECT content, message_type, photo_url FROM messages WHERE id = ?", originalMsgID).Scan(&content, &msgType, &photoURL)
	if err != nil { return err }

	// Invia come nuovo messaggio
	_, err = SendMessage(targetConversationID, newSenderID, content, msgType, photoURL, "")
	return err
}

// CommentMessage - Aggiunge una reazione
func CommentMessage(messageID string, userID string, emoji string) error {
	_, err := db.Exec("INSERT INTO reactions (message_id, user_id, emoji) VALUES (?, ?, ?)", messageID, userID, emoji)
	return err
}

// UncommentMessage - Rimuove una reazione
func UncommentMessage(messageID string, userID string) error {
	_, err := db.Exec("DELETE FROM reactions WHERE message_id = ? AND user_id = ?", messageID, userID)
	return err
}

// DeleteMessage - Cancella un messaggio (solo se inviato dall'utente)
func DeleteMessage(messageID string, userID string) error {
	res, err := db.Exec("DELETE FROM messages WHERE id = ? AND sender_id = ?", messageID, userID)
	if err != nil { return err }
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("message not found or not owned by user")
	}
	return nil
}

// --- Gruppi ---

// SetGroupName - Imposta nome gruppo (o crea il gruppo se usiamo un ID generato dal client)
func SetGroupName(groupID string, name string) error {
	// Usiamo INSERT ... ON DUPLICATE KEY UPDATE per gestire sia creazione che aggiornamento
	// Nota: is_group sarà true
	_, err := db.Exec(`
	INSERT INTO conversations (id, name, is_group, last_message_at)
	VALUES (?, ?, TRUE, NOW())
	ON DUPLICATE KEY UPDATE name = ?
	`, groupID, name, name)
	return err
}

// SetGroupPhoto - Imposta foto gruppo
func SetGroupPhoto(groupID string, photoURL string) error {
	_, err := db.Exec("UPDATE conversations SET photo_url = ? WHERE id = ?", photoURL, groupID)
	return err
}

// AddToGroup - Aggiunge utente al gruppo
func AddToGroup(groupID string, userID string) error {
	// Assicuriamoci che il gruppo esista e sia un gruppo
	var isGroup bool
	err := db.QueryRow("SELECT is_group FROM conversations WHERE id = ?", groupID).Scan(&isGroup)
	if err != nil { return errors.New("group not found") }
	if !isGroup { return errors.New("id refers to a direct conversation, not a group") }

	// Inserisci partecipante (ignorando duplicati)
	_, err = db.Exec("INSERT IGNORE INTO participants (conversation_id, user_id) VALUES (?, ?)", groupID, userID)
	return err
}

// LeaveGroup - Rimuove utente dal gruppo
func LeaveGroup(groupID string, userID string) error {
	_, err := db.Exec("DELETE FROM participants WHERE conversation_id = ? AND user_id = ?", groupID, userID)
	return err
}
