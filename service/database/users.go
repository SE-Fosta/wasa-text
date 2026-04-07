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
func (db *appdb) GetUsers(searchQuery string) ([]User, error) {
	var rows *sql.Rows
	var err error

	// Usiamo CAST per convertire l'INTEGER in stringa
	// Usiamo COALESCE per evitare i NULL e restituire "" se non c'è foto
	if searchQuery != "" {
		query := `SELECT CAST(id AS TEXT), username, COALESCE(photo_url, '') FROM users WHERE username LIKE ?`
		rows, err = db.c.Query(query, "%"+searchQuery+"%")
	} else {
		query := `SELECT CAST(id AS TEXT), username, COALESCE(photo_url, '') FROM users`
		rows, err = db.c.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		// ATTENZIONE: Qui usiamo u.PhotoURL con "URL" maiuscolo, esattamente come nella tua struct!
		if err := rows.Scan(&u.ID, &u.Username, &u.PhotoURL); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if users == nil {
		users = make([]User, 0)
	}

	return users, nil
}
