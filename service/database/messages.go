package database

import (
	"database/sql"
	"errors"
	"fmt"
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

type Reaction struct {
	UserID string `json:"userId"`
	Emoji  string `json:"emoji"`
}

func (db *appdb) SendMessage(conversationID string, senderID string, messageType string, content string, photoURL string, replyTo string) (Message, error) {
	// 1. Gestione del ReplyTo (NULL se vuoto)
	var replyToVal interface{} = replyTo
	if replyTo == "" {
		replyToVal = nil
	}

	// 2. Inserimento del messaggio nella tabella principale
	res, err := db.c.Exec(`
        INSERT INTO messages (conversation_id, sender_id, content, message_type, photo_url, reply_to) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		conversationID, senderID, content, messageType, photoURL, replyToVal)

	if err != nil {
		return Message{}, err
	}

	// 3. Recupero l'ID del messaggio appena creato
	lastID, _ := res.LastInsertId()
	// Lo teniamo come int64 per il database, ma lo convertiamo in string per la struct
	msgIDStr := strconv.FormatInt(lastID, 10)

	// 4. LOGICA DELLE SPUNTE: Creiamo le righe di "stato" per gli ALTRI membri
	// Inseriamo una riga in message_status per ogni membro della chat che NON è il mittente
	_, err = db.c.Exec(`
        INSERT INTO message_status (message_id, user_id, delivered, read)
        SELECT ?, user_id, 0, 0 
        FROM conversation_members 
        WHERE conversation_id = ? AND user_id != ?`,
		lastID, conversationID, senderID)

	if err != nil {
		// Se fallisce qui, meglio loggare l'errore ma non bloccare l'invio del messaggio
		fmt.Printf("Errore creazione message_status: %v\n", err)
	}

	// 5. Restituisco l'oggetto Message completo per il frontend
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
	// Usiamo sql.NullString anche per il content, per evitare crash se si inoltra una foto senza testo
	var content, photoURL sql.NullString

	// 1. Recuperiamo il contenuto e il tipo del messaggio originale
	err := db.c.QueryRow("SELECT content, message_type, photo_url FROM messages WHERE id = ?", messageID).
		Scan(&content, &msgType, &photoURL)
	if err != nil {
		return Message{}, err
	}

	// 2. Estraiamo i valori sicuri (se sono NULL nel DB, diventano stringhe vuote in Go "")
	safeContent := ""
	if content.Valid {
		safeContent = content.String
	}

	safePhotoURL := ""
	if photoURL.Valid {
		safePhotoURL = photoURL.String
	}

	// 3. Usiamo SendMessage per fare tutto il lavoro pesante!
	// Passiamo "" per replyTo poiché è un inoltro
	return db.SendMessage(targetConversationID, senderID, msgType, safeContent, safePhotoURL, "")
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

// GetMessages estrae tutti i messaggi di una specifica conversazione
func (db *appdb) GetMessages(conversationID string) ([]Message, error) {
	// 1. Aggiungiamo ms.delivered e ms.read alla SELECT
	// 2. Usiamo LEFT JOIN su message_status per non perdere i messaggi se lo stato manca
	query := `
        SELECT 
            m.id, m.content, m.message_type, m.timestamp, m.sender_id, 
            u.username as sender_name, m.reply_to, m.photo_url,
            COALESCE(MIN(ms.delivered), 0), 
            COALESCE(MIN(ms.read), 0)
        FROM messages m
        INNER JOIN users u ON m.sender_id = u.id
        LEFT JOIN message_status ms ON m.id = ms.message_id AND ms.user_id != m.sender_id
        WHERE m.conversation_id = ?
        GROUP BY m.id, m.content, m.message_type, m.timestamp, m.sender_id, u.username, m.reply_to, m.photo_url
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
		var content, replyTo, photoURL sql.NullString

		// Aggiungiamo &m.Delivered e &m.Read allo Scan
		// IMPORTANTE: L'ordine deve corrispondere esattamente alla SELECT sopra
		err := rows.Scan(
			&m.ID,
			&content,
			&m.MessageType,
			&m.Timestamp,
			&m.SenderID,
			&m.SenderName,
			&replyTo,
			&photoURL,
			&m.Delivered,
			&m.Read,
		)
		if err != nil {
			return nil, err
		}

		// Mantieni tutte le tue logiche esistenti per i campi Null
		if content.Valid {
			m.Content = content.String
		}
		if replyTo.Valid {
			m.ReplyTo = replyTo.String
		}
		if photoURL.Valid {
			m.PhotoURL = photoURL.String
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
		if err == nil {
			defer rRows.Close()
			reactionsMap := make(map[string][]Reaction)

			for rRows.Next() {
				var msgID, userID, emoji string
				if rRows.Scan(&msgID, &userID, &emoji) == nil {
					reactionsMap[msgID] = append(reactionsMap[msgID], Reaction{
						UserID: userID,
						Emoji:  emoji,
					})
				}
			}

			for i, msg := range messages {
				if rx, found := reactionsMap[msg.ID]; found {
					messages[i].Reactions = rx
				}
			}
		}
	}

	return messages, nil
}
func (db *appdb) MarkAsRead(conversationID string, userID string) error {
	// Questa query aggiorna lo stato per l'utente corrente
	// su tutti i messaggi di quella conversazione che non ha ancora letto
	query := `
        UPDATE message_status 
        SET read = 1, delivered = 1 
        WHERE user_id = ? 
        AND message_id IN (
            SELECT id FROM messages WHERE conversation_id = ? AND sender_id != ?
        )`

	_, err := db.c.Exec(query, userID, conversationID, userID)
	if err != nil {
		fmt.Printf("--- ERRORE SQL MarkAsRead: %v ---\n", err)
		return err
	}
	return nil
}

// Aggiunge o aggiorna una reazione (emoji) di un utente a un messaggio.
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

// Rimuove la reazione di un utente da un messaggio specifico.
func (db *appdb) UnreactMessage(messageID string, userID string) error {
	query := `
		DELETE FROM reactions 
		WHERE message_id = ? AND user_id = ?
	`
	_, err := db.c.Exec(query, messageID, userID)
	return err
}

// CommentMessage crea una risposta a un messaggio esistente
func (db *appdb) CommentMessage(originalMessageID string, senderID string, content string) (Message, error) {
	// 1. Per prima cosa, dobbiamo capire a quale conversazione appartiene il messaggio originale
	var conversationID string
	err := db.c.QueryRow("SELECT conversation_id FROM messages WHERE id = ?", originalMessageID).Scan(&conversationID)
	if err != nil {
		return Message{}, err // Il messaggio originale non esiste
	}

	// 2. Ora possiamo usare la tua funzione SendMessage per fare l'inserimento!
	// Passiamo originalMessageID come ultimo parametro (replyTo)
	return db.SendMessage(conversationID, senderID, "text", content, "", originalMessageID)
}

// UncommentMessage elimina un commento (che di fatto è un messaggio)
func (db *appdb) UncommentMessage(commentID string, requestingUserID string) error {
	// Poiché il commento è salvato nella tabella messages, usiamo la tua DeleteMessage
	return db.DeleteMessage(commentID, requestingUserID)
}
