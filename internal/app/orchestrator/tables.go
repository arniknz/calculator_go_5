package application

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func NewDB() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", "calculator_5_go.db")
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	schema := `
				CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				login TEXT UNIQUE NOT NULL,
				password TEXT NOT NULL
				);
				CREATE TABLE IF NOT EXISTS expressions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				expression TEXT NOT NULL,
				status TEXT NOT NULL,
				result REAL,
				FOREIGN KEY(user_id) REFERENCES users(id)
				);
	`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal("migrate failed:", err)
	}
	return db
}
