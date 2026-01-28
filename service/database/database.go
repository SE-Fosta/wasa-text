package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings" // <--- IMPORTANTE: AGGIUNGI QUESTO IMPORT

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func InitDatabase(dataSourceName string) (*sql.DB, error) {
	var err error
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Database connected successfully")

	if err := createTables(); err != nil {
		return nil, fmt.Errorf("error creating tables: %w", err)
	}

	return db, nil
}

func createTables() error {
	path := "service/database/schema.sql"

	content, err := ioutil.ReadFile(path)
	if err != nil {
		path = "../../service/database/schema.sql"
		content, err = ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("impossibile leggere schema.sql: %w", err)
		}
	}

	sqlStatements := string(content)

	// --- MODIFICA QUI: Dividiamo le query per punto e virgola ---
	queries := strings.Split(sqlStatements, ";")

	for _, query := range queries {
		// Puliamo la query da spazi e a capo inutili
		cleanQuery := strings.TrimSpace(query)
		
		// Se la query è vuota (es. l'ultimo punto e virgola del file), la saltiamo
		if cleanQuery == "" {
			continue
		}

		// Eseguiamo la singola query
		_, err = db.Exec(cleanQuery)
		if err != nil {
			return fmt.Errorf("errore esecuzione query SQL: %s \n Errore: %w", cleanQuery, err)
		}
	}

	log.Println("Tables created/verified successfully")
	return nil
}