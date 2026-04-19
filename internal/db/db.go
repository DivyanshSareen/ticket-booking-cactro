package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(path string) {
	var err error
	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	DB.Exec("PRAGMA journal_mode=WAL;")
	DB.Exec("PRAGMA foreign_keys=ON;")
	migrate()
}

func migrate() {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		email        TEXT    NOT NULL UNIQUE,
		password_hash TEXT   NOT NULL,
		role         TEXT    NOT NULL CHECK(role IN ('organizer','customer')),
		created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS events (
		id                INTEGER PRIMARY KEY AUTOINCREMENT,
		organizer_id      INTEGER NOT NULL REFERENCES users(id),
		name              TEXT    NOT NULL,
		date              TEXT    NOT NULL,
		location          TEXT    NOT NULL,
		total_tickets     INTEGER NOT NULL DEFAULT 50,
		available_tickets INTEGER NOT NULL DEFAULT 50,
		created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS bookings (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		customer_id INTEGER NOT NULL REFERENCES users(id),
		event_id    INTEGER NOT NULL REFERENCES events(id),
		num_tickets INTEGER NOT NULL CHECK(num_tickets > 0),
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := DB.Exec(schema); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}
