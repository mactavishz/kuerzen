package url

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type URLStore interface {
	CreateShortURL(string, string, context.Context) error
	GetLongURL(string, context.Context) (string, error)
}

type PostgresURLStore struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewPostgresURLStore(db *sql.DB, logger *zap.SugaredLogger) *PostgresURLStore {
	return &PostgresURLStore{
		db:     db,
		logger: logger,
	}
}

var ErrDuplicateLongURL = errors.New("long URL already exists")
var ErrShortURLNotFound = errors.New("short URL not found")

func (pgs *PostgresURLStore) CreateShortURL(shortURL string, longURL string, ctx context.Context) error {
	operation := func() error {
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		tx, err := pgs.db.BeginTx(dbCtx, nil)
		if err != nil {
			pgs.logger.Errorf("Failed to begin transaction for short URL %s: %v", shortURL, err)
			return err
		}
		defer tx.Rollback()
		query := `
		INSERT INTO urls(short_url, long_url)
		VALUES($1, $2)
		`
		_, err = tx.ExecContext(dbCtx, query, shortURL, longURL)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
				pgs.logger.Debugf("Attempted to create duplicate long URL: %s", longURL)
				return backoff.Permanent(ErrDuplicateLongURL)
			}
			pgs.logger.Errorf("Failed to insert short URL %s: %v", shortURL, err)
			return err
		}
		err = tx.Commit()
		if err != nil {
			pgs.logger.Errorf("Failed to commit transaction for short URL %s: %v", shortURL, err)
			return err
		}
		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second
	b.InitialInterval = 50 * time.Millisecond
	b.MaxInterval = 5 * time.Second
	b.RandomizationFactor = 0.5

	err := backoff.Retry(operation, backoff.WithContext(b, ctx))

	if err != nil {
		if errors.Is(err, ErrDuplicateLongURL) {
			return ErrDuplicateLongURL
		}
		return err
	}

	pgs.logger.Infow("Successfully created short URL", "shortURL", shortURL, "longURL", longURL)
	return nil
}

func (pgs *PostgresURLStore) GetLongURL(shortURL string, ctx context.Context) (string, error) {
	var longURL string
	query := `
	SELECT long_url from urls
	WHERE short_url = $1
	`
	operation := func() error {
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		err := pgs.db.QueryRowContext(dbCtx, query, shortURL).Scan(&longURL)
		if err != nil {
			if err == sql.ErrNoRows {
				pgs.logger.Debugf("Short URL %s not found in DB, no retry.", shortURL)
				return backoff.Permanent(ErrShortURLNotFound)
			}
			pgs.logger.Warnf("Attempt to get long URL for %s from DB failed, retrying: %v", shortURL, err)
			return err
		}
		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second
	b.InitialInterval = 50 * time.Millisecond
	b.MaxInterval = 5 * time.Second
	b.RandomizationFactor = 0.5

	err := backoff.Retry(operation, backoff.WithContext(b, ctx))

	if err != nil {
		if errors.Is(err, ErrShortURLNotFound) {
			return "", ErrShortURLNotFound
		}
		return "", err
	}
	pgs.logger.Debugf("Successfully retrieved long URL for %s on attempt.", shortURL)
	return longURL, nil
}
