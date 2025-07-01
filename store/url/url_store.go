package url

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"math/rand"
	"time"

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
var ErrMaxElapsedTimeExceeded = errors.New("max elapsed time for operation exceeded")
var ErrRetriesExhausted = errors.New("operation failed after all retries exhausted")

const (
	maxElapsedTime       = 30 * time.Second      //Maximum total duration that all retries may last
	initialSleepInterval = 50 * time.Millisecond //The initial waiting time before the first retry
	maxSleepInterval     = 5 * time.Second       //The longest possible waiting time between two retries
	maxRetries           = 10                    //Upper limit for the number of retries if time is not the primary termination condition
)

func (pgs *PostgresURLStore) CreateShortURL(shortURL string, longURL string, ctx context.Context) error {
	query := `
		INSERT INTO urls(short_url, long_url)
		VALUES($1, $2)
		`
	startTime := time.Now()
	sleep := initialSleepInterval
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			pgs.logger.Infof("CreateShortURL operation cancelled for %s: %v", shortURL, ctx.Err())
			return ctx.Err()
		default:
		}
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		tx, err := pgs.db.BeginTx(dbCtx, nil)
		if err != nil {
			pgs.logger.Errorf("Failed to begin transaction for short URL %s: %v", shortURL, err)
			goto retryAttempt
		}
		defer tx.Rollback()
		_, err = tx.ExecContext(dbCtx, query, shortURL, longURL)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
				pgs.logger.Infof("Attempted to create duplicate long URL: %s", longURL)
				return ErrDuplicateLongURL
			}
			pgs.logger.Errorf("Failed to insert short URL %s: %v", shortURL, err)
			goto retryAttempt
		}
		err = tx.Commit()
		if err != nil {
			pgs.logger.Warnf("Attempt to commit transaction for short URL %s failed, retrying: %v", shortURL, err)
			goto retryAttempt
		}
		pgs.logger.Infow("Successfully created short URL", "shortURL", shortURL, "longURL", longURL)
		return nil
	retryAttempt:
		//sleep = min(cap, random_between(base, sleep * 3)
		sleep = time.Duration(math.Min(float64(maxSleepInterval), float64(initialSleepInterval)+rand.Float64()*float64(3*sleep-initialSleepInterval)))
		pgs.logger.Infof("Waiting for %v before next retry attempt for %s (attempt %d)", sleep, shortURL, i+1)
		//Proactive check whether the maxSleepInterval would be exceeded after waiting sleep-long
		if time.Since(startTime)+sleep >= maxElapsedTime {
			pgs.logger.Infof("CreateShortURL operation timed out after %v for %s", maxElapsedTime, shortURL)
			return ErrMaxElapsedTimeExceeded
		}
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			pgs.logger.Infof("CreateShortURL operation cancelled during backoff for %s: %v", shortURL, ctx.Err())
			return ctx.Err()
		}
	}
	pgs.logger.Infof("failed to create short URL after multiple retries")
	return ErrRetriesExhausted
}

func (pgs *PostgresURLStore) GetLongURL(shortURL string, ctx context.Context) (string, error) {
	var longURL string
	query := `
	SELECT long_url from urls
	WHERE short_url = $1
	`
	startTime := time.Now()
	sleep := initialSleepInterval
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			pgs.logger.Infof("GetLongURL operation cancelled for %s: %v", shortURL, ctx.Err())
			return "", ctx.Err()
		default:
		}
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		err := pgs.db.QueryRowContext(dbCtx, query, shortURL).Scan(&longURL)
		if err != nil {
			if err == sql.ErrNoRows {
				pgs.logger.Debugf("Short URL %s not found in DB, no retry.", shortURL)
				return "", ErrShortURLNotFound // Permanent error, no retry
			}
			pgs.logger.Warnf("Attempt to get long URL for %s from DB failed, retrying: %v", shortURL, err)
			goto retryAttemptGetLongURL
		}
		pgs.logger.Debugf("Successfully retrieved long URL for %s on attempt.", shortURL)
		return longURL, nil

	retryAttemptGetLongURL:
		sleep = time.Duration(math.Min(float64(maxSleepInterval), float64(initialSleepInterval)+rand.Float64()*float64(3*sleep-initialSleepInterval)))
		pgs.logger.Infof("Waiting for %v before next retry attempt for %s (attempt %d)", sleep, shortURL, i+1)

		if time.Since(startTime)+sleep >= maxElapsedTime {
			pgs.logger.Infof("GetLongURL operation timed out after %v for %s", maxElapsedTime, shortURL)
			return "", ErrMaxElapsedTimeExceeded
		}
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			pgs.logger.Infof("GetLongURL operation cancelled during backoff for %s: %v", shortURL, ctx.Err())
			return "", ctx.Err()
		}
	}
	pgs.logger.Infof("Failed to get long URL after multiple retries for %s", shortURL)
	return "", ErrRetriesExhausted
}
