package url

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mactavishz/kuerzen/retries"
	"go.uber.org/zap"
)

type URLStore interface {
	CreateShortURL(string, string, context.Context) func() retries.RetryableFuncObject
	GetLongURL(string, context.Context) func() retries.RetryableFuncObject
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

func (pgs *PostgresURLStore) CreateShortURL(shortURL string, longURL string, ctx context.Context) func() retries.RetryableFuncObject {
	query := `
		INSERT INTO urls(short_url, long_url)
		VALUES($1, $2)
		`
	return func() retries.RetryableFuncObject {
		var rfo retries.RetryableFuncObject
		rfo.Ctx = ctx
		rfo.Logger = pgs.logger
		select {
		case <-ctx.Done():
			pgs.logger.Infof("CreateShortURL operation cancelled for %s: %v", shortURL, ctx.Err())
			rfo.Err = ctx.Err()
			return rfo
		default:
		}
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		tx, err := pgs.db.BeginTx(dbCtx, nil)
		if err != nil {
			pgs.logger.Errorf("Failed to begin transaction for short URL %s: %v", shortURL, err)
			rfo.Err = retries.ErrTransient
			return rfo
		}
		defer tx.Rollback()
		_, err = tx.ExecContext(dbCtx, query, shortURL, longURL)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
				pgs.logger.Errorf("Attempted to create duplicate long URL: %s", longURL)
				rfo.Err = ErrDuplicateLongURL
				return rfo
			}
			pgs.logger.Errorf("Failed to insert short URL %s: %v", shortURL, err)
			rfo.Err = retries.ErrTransient
			return rfo
		}
		err = tx.Commit()
		if err != nil {
			pgs.logger.Infof("Attempt to commit transaction for short URL %s failed, retrying: %v", shortURL, err)
			rfo.Err = retries.ErrTransient
			return rfo
		}
		pgs.logger.Infow("Successfully created short URL", "shortURL", shortURL, "longURL", longURL)
		rfo.Err = nil
		return rfo
	}
}

func (pgs *PostgresURLStore) GetLongURL(shortURL string, ctx context.Context) func() retries.RetryableFuncObject {
	var longURL string
	query := `
	SELECT long_url from urls
	WHERE short_url = $1
	`
	return func() retries.RetryableFuncObject {
		var rfo retries.RetryableFuncObject
		rfo.Ctx = ctx
		rfo.Logger = pgs.logger
		select {
		case <-ctx.Done():
			pgs.logger.Infof("GetLongURL operation cancelled for %s: %v", shortURL, ctx.Err())
			rfo.Err = ctx.Err()
			rfo.Rest = append(rfo.Rest, "")
			return rfo
		default:
		}
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		err := pgs.db.QueryRowContext(dbCtx, query, shortURL).Scan(&longURL)
		if err != nil {
			if err == sql.ErrNoRows {
				pgs.logger.Errorf("Short URL %s not found in DB, no retry.", shortURL)
				rfo.Err = ErrShortURLNotFound
				rfo.Rest = append(rfo.Rest, "")
				return rfo
			}
			pgs.logger.Infof("Attempt to get long URL for %s from DB failed, retrying: %v", shortURL, err)
			rfo.Err = retries.ErrTransient
			rfo.Rest = append(rfo.Rest, "")
			return rfo
		}
		pgs.logger.Infof("Successfully retrieved long URL for %s on attempt.", shortURL)
		rfo.Err = nil
		rfo.Rest = append(rfo.Rest, longURL)
		return rfo
	}
}
