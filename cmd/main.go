package main

import (
	"database/sql"
	"log"
)

func main() {
	db, err := sql.Open("postgres", "host=postgres port=5432 user=order_user password=order_password dbname=orders_db sslmode=disable")
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer db.Close()
}
