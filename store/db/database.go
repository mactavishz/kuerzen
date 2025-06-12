package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Opens a connection to database use pgx as the driver
// The returned [DB] is safe for concurrent use by multiple goroutines (maintains a pool of connection)
func Open() (*sql.DB, error) {
	db, err := sql.Open("pgx", os.Getenv("KUERZEN_DB_URL"))

	if err != nil {
		return nil, fmt.Errorf("db: open %w", err)
	}

	log.Println("Connected to the database")

	return db, nil
}

func MigrateFS(db *sql.DB, migrationFS fs.FS, dir string) error {
	goose.SetBaseFS(migrationFS)
	defer func() {
		goose.SetBaseFS(nil)
	}()
	return migrate(db, dir)
}

// run database migration
func migrate(db *sql.DB, dir string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	err = goose.Up(db, dir)
	if err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
