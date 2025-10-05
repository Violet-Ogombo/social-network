package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	// Register the file source for golang-migrate (allows file:// URLs)
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
)

var DB *sql.DB

// requiredTables lists tables we expect after migrations
var requiredTables = []string{"users", "sessions", "messages"}

func InitDB() {
	var err error

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./backend/socialnetwork.db"
	}

	// convert to absolute path for clarity
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve DB path: %v", err))
	}
	log.Printf("Opening SQLite DB at %s", absPath)

	DB, err = sql.Open("sqlite3", absPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to DB: %v", err))
	}

	// Run migrations
	driver, err := sqlite3.WithInstance(DB, &sqlite3.Config{})
	if err != nil {
		panic(fmt.Sprintf("Migration driver error: %v", err))
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://backend/db/migrations/sqlite",
		"sqlite3", driver)
	if err != nil {
		panic(fmt.Sprintf("Migration setup error: %v", err))
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(fmt.Sprintf("Migration failed: %v", err))
	}

	// simple schema check: ensure required tables exist
	for _, t := range requiredTables {
		var name string
		row := DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", t)
		if err := row.Scan(&name); err != nil {
			log.Printf("Warning: expected table '%s' not found in DB (%s)", t, absPath)
		}
	}
}
