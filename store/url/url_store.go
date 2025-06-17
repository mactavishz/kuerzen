package url

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type URLStore interface {
	CreateShortURL(string, string) error
	GetLongURL(string) (string, error)
}

type PostgresURLStore struct {
	db *sql.DB
}

func NewPostgresURLStore(db *sql.DB) *PostgresURLStore {
	return &PostgresURLStore{
		db: db,
	}
}

var ErrDuplicateLongURL = errors.New("long URL already exists")
var ErrShortURLNotFound = errors.New("short URL not found")

func (pgs *PostgresURLStore) CreateShortURL(shortURL string, longURL string) error {
	tx, err := pgs.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	INSERT INTO urls(short_url, long_url)
	VALUES($1, $2)
	`
	_, err = tx.Exec(query, shortURL, longURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return ErrDuplicateLongURL
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (pgs *PostgresURLStore) GetLongURL(shortURL string) (string, error) {
	var longURL string
	query := `
	SELECT long_url from urls
	WHERE short_url = $1
	`
	err := pgs.db.QueryRow(query, shortURL).Scan(&longURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrShortURLNotFound
		}
		return "", err
	}
	return longURL, nil
}
