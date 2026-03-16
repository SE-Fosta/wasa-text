package database

import (
	"database/sql"
	"errors"
	"strconv"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	PhotoURL string `json:"photoUrl,omitempty"`
}

func (db *appdb) DoLogin(username string) (string, error) {
	var idInt int64
	err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&idInt)
	
	if errors.Is(err, sql.ErrNoRows) {
		// Utente non esiste, lo inseriamo
		res, err := db.c.Exec("INSERT INTO users (username) VALUES (?)", username)
		if err != nil {
			return "", err
		}
		idInt, err = res.LastInsertId()
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(idInt, 10), nil
	} else if err != nil {
		return "", err
	}
	// Utente esiste
	return strconv.FormatInt(idInt, 10), nil
}

func (db *appdb) SetMyUserName(userID string, newName string) error {
	res, err := db.c.Exec("UPDATE users SET username = ? WHERE id = ?", newName, userID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("user not found")
	}
	return err
}

func (db *appdb) SetMyPhoto(userID string, photoURL string) error {
	_, err := db.c.Exec("UPDATE users SET photo_url = ? WHERE id = ?", photoURL, userID)
	return err
}