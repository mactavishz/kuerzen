package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

type Database struct {
	DB     *sql.DB
	logger *zap.SugaredLogger
}

func NewDatabase(logger *zap.SugaredLogger) *Database {
	return &Database{
		logger: logger,
	}
}

// Opens a connection to database use pgx as the driver
// The returned [DB] is safe for concurrent use by multiple goroutines (maintains a pool of connection)
func (d *Database) Open() error {
	db, err := sql.Open("pgx", os.Getenv("KUERZEN_DB_URL"))

	if err != nil {
		return fmt.Errorf("open: %w", err)
	}

	d.DB = db
	d.logger.Infof("Connected to the database")

	return nil
}

func (d *Database) MigrateFS(migrationFS fs.FS, dir string) error {
	goose.SetBaseFS(migrationFS)
	defer func() {
		goose.SetBaseFS(nil)
	}()
	return d.migrate(dir)
}

// run database migration
func (d *Database) migrate(dir string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	err = goose.Up(d.DB, dir)
	if err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func (d *Database) Close() error {
	if d.DB == nil {
		return nil
	}

	err := d.DB.Close()
	if err != nil {
		return fmt.Errorf("close: %w", err)
	}

	d.logger.Infof("Database connection closed")
	return nil
}
