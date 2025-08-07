package postgres

import (
	"database/sql"
	"fmt"
)

func ConnectDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		defer func() {
			if err := db.Close(); err != nil {
				fmt.Println("failed to close db", err)
			}
		}()
		return nil, err
	}

	return db, nil
}
