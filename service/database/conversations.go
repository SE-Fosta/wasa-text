package database

import (
	"database/sql"
	"errors"
	"fmt"
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
	// Abbiamo trasformato la query per usare delle JOIN molto più efficienti
	// e abbiamo aggiunto il calcolo dei messaggi non letti (UnreadCount)
	query := `
        SELECT 
            c.id, 
            CASE 
                WHEN c.is_group = 1 THEN c.name 
                ELSE (
                    SELECT u.username 
                    FROM conversation_members m 
                    INNER JOIN users u ON m.user_id = u.id 
                    WHERE m.conversation_id = c.id AND u.id != ?
                ) 
            END as chat_name,
            c.photo_url, 
            c.is_group,
            lm.content as last_msg_content,
            lm.message_type as last_msg_type,
            lu.username as last_msg_sender,
            lm.timestamp as last_timestamp,
            (
                SELECT COUNT(*) 
                FROM message_status ms 
                INNER JOIN messages m ON ms.message_id = m.id 
                WHERE m.conversation_id = c.id AND ms.user_id = ? AND ms.read = 0
            ) as unread_count
        FROM conversations c
        INNER JOIN conversation_members cm ON c.id = cm.conversation_id
        -- Trucco SQL: Troviamo l'ID del messaggio più recente per ogni conversazione
        LEFT JOIN (
            SELECT conversation_id, MAX(id) as max_msg_id
            FROM messages
            GROUP BY conversation_id
        ) as latest_msg ON c.id = latest_msg.conversation_id
        -- Uniamo i dettagli di quell'ultimo messaggio
        LEFT JOIN messages lm ON latest_msg.max_msg_id = lm.id
        -- Uniamo l'utente che ha inviato l'ultimo messaggio (per il SenderName)
        LEFT JOIN users lu ON lm.sender_id = lu.id
        WHERE cm.user_id = ?
        AND latest_msg.max_msg_id IS NOT NULL -- Esclude le chat vuote
        ORDER BY last_timestamp DESC
    `

	// ATTENZIONE: Ora ci sono tre '?' nella query, quindi passiamo userID 3 volte!
	rows, err := db.c.Query(query, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []ConversationSummary
	for rows.Next() {
		var s ConversationSummary
		var idInt int64

		// Prepariamo le variabili "sicure" per i dati che potrebbero essere nulli
		var name, photoURL, lastMsgContent, lastMsgType, lastMsgSender sql.NullString
		var lastTimestamp sql.NullTime
		var unreadCount int

		// L'ordine DEVE essere identico alla SELECT in alto
		err := rows.Scan(
			&idInt,
			&name,
			&photoURL,
			&s.IsGroup,
			&lastMsgContent,
			&lastMsgType,
			&lastMsgSender,
			&lastTimestamp,
			&unreadCount,
		)
		if err != nil {
			return nil, err
		}

		// Assegnazioni Base
		s.ID = strconv.FormatInt(idInt, 10)
		if name.Valid {
			s.Name = name.String
		}
		if photoURL.Valid {
			s.PhotoURL = photoURL.String
		}
		s.UnreadCount = unreadCount

		// --- POPOLIAMO LA NUOVA STRUCT ANNIDATA (MessagePreview) ---
		s.LastMessage = MessagePreview{}

		if lastMsgContent.Valid && lastMsgContent.String != "" {
			s.LastMessage.Content = lastMsgContent.String
		} else {
			s.LastMessage.Content = "📷 Foto" // Testo di fallback se è una foto
		}

		if lastMsgType.Valid {
			s.LastMessage.MessageType = lastMsgType.String
		}

		if lastMsgSender.Valid {
			s.LastMessage.SenderName = lastMsgSender.String
		}

		// Gestione Timestamp
		if lastTimestamp.Valid {
			s.LastActivity = lastTimestamp.Time
		} else {
			s.LastActivity = time.Now()
		}

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
func (db *appdb) CreateConversation(creatorID string, targetUserID string, isGroup bool, groupName string) (string, error) {
	// ==========================================
	// SCENARIO 1: CREAZIONE DI UN GRUPPO
	// ==========================================
	if isGroup {
		// 1. Creiamo la conversazione impostando is_group a 1 e salvando il nome
		res, err := db.c.Exec(`INSERT INTO conversations (is_group, name, photo_url) VALUES (1, ?, '')`, groupName)
		if err != nil {
			return "", err
		}

		convID, err := res.LastInsertId()
		if err != nil {
			return "", err
		}

		// 2. Aggiungiamo il creatore come PRIMO membro del gruppo
		_, err = db.c.Exec(`INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`, convID, creatorID)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%d", convID), nil
	}

	// ==========================================
	// SCENARIO 2: CREAZIONE CHAT 1-A-1 (La tua logica intatta)
	// ==========================================
	checkQuery := `
        SELECT c.id FROM conversations c
        JOIN conversation_members m1 ON c.id = m1.conversation_id
        JOIN conversation_members m2 ON c.id = m2.conversation_id
        WHERE c.is_group = 0 AND m1.user_id = ? AND m2.user_id = ?
    `
	var existingConvID int64
	err := db.c.QueryRow(checkQuery, creatorID, targetUserID).Scan(&existingConvID)

	if err == nil {
		// La conversazione esiste già! Restituiamo il suo ID senza creare doppioni
		return fmt.Sprintf("%d", existingConvID), nil
	}

	// 2. Se non esiste, creiamo la nuova riga in 'conversations'
	res, err := db.c.Exec(`INSERT INTO conversations (is_group, name, photo_url) VALUES (0, '', '')`)
	if err != nil {
		return "", err
	}

	convID, err := res.LastInsertId()
	if err != nil {
		return "", err
	}

	// 3. Aggiungiamo i due utenti come membri della conversazione
	_, err = db.c.Exec(`INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`, convID, creatorID)
	if err != nil {
		return "", err
	}

	_, err = db.c.Exec(`INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`, convID, targetUserID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", convID), nil
}

func (db *appdb) AddToGroup(groupID string, userIDToAdd string) error {
	// 1. Controlliamo esplicitamente: l'utente c'è già?
	var exists bool
	err := db.c.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM conversation_members 
            WHERE conversation_id = ? AND user_id = ?
        )`, groupID, userIDToAdd).Scan(&exists)

	if err != nil {
		return err // Vero guasto del DB
	}

	if exists {
		// L'utente c'è già! Segnaliamo all'API di fermarsi.
		return errors.New("user is already in the group")
	}

	// 2. Se non c'è, procediamo con l'inserimento normale
	_, err = db.c.Exec("INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)", groupID, userIDToAdd)
	return err
}

func (db *appdb) LeaveGroup(groupID string, userIDToRemove string) error {
	// Rimuoviamo il record dalla tabella dei membri
	_, err := db.c.Exec("DELETE FROM conversation_members WHERE conversation_id = ? AND user_id = ?", groupID, userIDToRemove)
	return err
}

func (db *appdb) SetGroupName(groupID string, newName string) error {
	// Aggiorniamo il nome nella tabella conversations
	_, err := db.c.Exec("UPDATE conversations SET name = ? WHERE id = ?", newName, groupID)
	return err
}

func (db *appdb) SetGroupPhoto(groupID string, photoURL string) error {
	_, err := db.c.Exec("UPDATE conversations SET photo_url = ? WHERE id = ?", photoURL, groupID)
	return err
}
