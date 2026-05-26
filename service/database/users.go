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

func (db *appdb) DoLogin(username string) (string, string, error) {
	var idInt int64
	var photoUrl sql.NullString

	query := "SELECT id, photo_url FROM users WHERE username = ?"
	err := db.c.QueryRow(query, username).Scan(&idInt, &photoUrl)

	if errors.Is(err, sql.ErrNoRows) {
		res, err := db.c.Exec("INSERT INTO users (username) VALUES (?)", username)
		if err != nil {
			return "", "", err
		}

		idInt, err = res.LastInsertId()
		if err != nil {
			return "", "", err
		}

		return strconv.FormatInt(idInt, 10), "", nil

	} else if err != nil {
		return "", "", err
	}

	return strconv.FormatInt(idInt, 10), photoUrl.String, nil
}

func (db *appdb) SetMyUserName(userID string, newName string) error {
	res, err := db.c.Exec("UPDATE users SET username = ? WHERE id = ?", newName, userID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
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

func (db *appdb) GetUser(userID string) (User, error) {
	var u User

	err := db.c.QueryRow(`
        SELECT id, name, IFNULL(photo_url, '') 
        FROM users 
        WHERE id = ?`, userID).Scan(&u.ID, &u.Username, &u.PhotoURL)

	if err != nil {
		return u, err
	}

	return u, nil
}
