package database

// Reaction rappresenta una singola emoji messa da un utente a un messaggio.
// I tag `json:"..."` servono per formattare correttamente la risposta per il frontend.
type Reaction struct {
	UserID string `json:"userId"`
	Emoji  string `json:"emoji"`
}

// CommentMessage aggiunge o aggiorna una reazione (emoji) di un utente a un messaggio.
func (db *appdb) CommentMessage(messageID string, userID string, emoji string) error {
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

// UncommentMessage rimuove la reazione di un utente da un messaggio specifico.
func (db *appdb) UncommentMessage(messageID string, userID string) error {
	query := `
		DELETE FROM reactions 
		WHERE message_id = ? AND user_id = ?
	`
	_, err := db.c.Exec(query, messageID, userID)
	return err
}