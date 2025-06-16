package url

import "database/sql"

type URLStore interface {
	Shorten(string) error
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

func (pgs *PostgresURLStore) CreateShortURL(shortURL string) error {
	panic("TODO")
}

func (pgs *PostgresURLStore) GetLongURL(longUrl string) (string, error) {
	panic("TODO")
}
