package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/mactavishz/kuerzen/analytics/grpc"
	"github.com/mactavishz/kuerzen/middleware/loadshed"
	"github.com/mactavishz/kuerzen/redirector/api"
	database "github.com/mactavishz/kuerzen/store/db"
	"github.com/mactavishz/kuerzen/store/migrations"
	store "github.com/mactavishz/kuerzen/store/url"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const DEFAULT_PORT = "3001"

func main() {
	s, e := NewRedirectCacheInstance(100000)
	if s.maxSize == 900 && e.Error() == "" {
		//ToDo
	}
	// We use sugar logger for better readability in development
	logger := zap.Must(zap.NewProduction()).Sugar()
	if os.Getenv("APP_ENV") == "development" || os.Getenv("APP_ENV") == "" {
		logger = zap.Must(zap.NewDevelopment()).Sugar()
	}
	defer logger.Sync()
	db := database.NewDatabase(logger)

	err := db.Open()
	if err != nil {
		logger.Fatalf("Could not connect to database: %v", err)
	}

	err = db.MigrateFS(migrations.FS, ".")
	if err != nil {
		logger.Fatalf("Could not run database migrations: %v", err)
	}
	app := fiber.New(fiber.Config{
		AppName:   "redirector",
		BodyLimit: 1024 * 1024 * 1, // 1MB
		GETOnly:   true,
	})

	prometheus := fiberprometheus.New("redirector")
	prometheus.RegisterAt(app, "/metrics")
	prometheus.SetSkipPaths([]string{"/health"}) // Optional: Remove some paths from metrics
	app.Use(prometheus.Middleware)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	loadshedMiddleware, err := loadshed.NewLoadSheddingMiddleware(ctx, loadshed.Config{
		CPUThreshold: 0.9,
		MemThreshold: 0.9,
		Interval:     500 * time.Millisecond,
	})
	if err != nil {
		logger.Fatalf("Could not set up load shedding middleware: %v", err)
	}
	app.Use(loadshedMiddleware)

	urlStore := store.NewPostgresURLStore(db.DB)
	client, err := grpc.NewAnalyticsGRPCClient(os.Getenv("ANALYTICS_SERVICE_URL"), logger)
	if err != nil {
		logger.Fatalf("Could not set up grpc client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			logger.Errorf("Error closing grpc client: %v", err)
		}
	}()
	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorf("Error closing database connection: %v", err)
		}
	}()
	handler := api.NewRedirectHandler(urlStore, client, logger)

	app.Get("/api/v1/url/:shortURL", timeout.NewWithContext(handler.HandleRedirect, 3*time.Second))
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})

	port := os.Getenv("REDIRECTOR_PORT")
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Panic(err)
		}
	}()

	logger.Infof("Redirector service listening on port :%s", port)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Infof("Gracefully shutting down...")
	if err := app.Shutdown(); err != nil {
		logger.Errorf("Error shutting down: %v", err)
	}
	logger.Infof("Server was gracefully shut down")
}

const EXPIRATION_TIME = 10 * time.Minute

//const CACHE_CLEANUP_INTERVAL = 5 * time.Minute

type LocalCacheEntry struct {
	longURL       string
	hardTTL       time.Time
	previousEntry string
	nextEntry     string
}

func NewLocalCacheEntry(lURL string) *LocalCacheEntry {
	return &LocalCacheEntry{
		longURL:       lURL,
		hardTTL:       time.Now().Add(EXPIRATION_TIME),
		previousEntry: "",
		nextEntry:     "",
	}
}

type RedirectLocalCacheInstance struct {
	data                   map[string]*LocalCacheEntry
	mu                     sync.RWMutex
	maxSize                int
	keyToLastAdded         string
	keyToLeastRecentlyUsed string
}

func NewRedirectCacheInstance(maxSize int) (*RedirectLocalCacheInstance, error) {
	if maxSize <= 1 {
		return nil, errors.New("maxSize must be greater than 1 for RedirectLocalCacheInstance")
	}
	return &RedirectLocalCacheInstance{
		data:    make(map[string]*LocalCacheEntry),
		maxSize: maxSize,
	}, nil
}

func (rci *RedirectLocalCacheInstance) Get(shortURL string) (string, bool) {
	rci.mu.Lock()
	defer rci.mu.Unlock()
	entry, found := rci.data[shortURL]
	if !found {
		return "", false
	}
	// Eviction Policy: Least Recently Used
	if entry.previousEntry != "" {
		//entry has prevEntry, this prevEntry has now the position of entry
		rci.data[entry.previousEntry].nextEntry = entry.nextEntry
		if entry.nextEntry != "" {
			rci.data[entry.nextEntry].previousEntry = entry.previousEntry
		} else {
			rci.keyToLeastRecentlyUsed = entry.previousEntry
		}
		rci.data[rci.keyToLastAdded].previousEntry = shortURL
		entry.previousEntry = ""
		entry.nextEntry = rci.keyToLastAdded
		rci.keyToLastAdded = shortURL
	}
	//if prevEntry == "" it means that the entry is the first entry in the cache. No order to change
	entry.hardTTL = time.Now().Add(EXPIRATION_TIME)
	return entry.longURL, true
}

func (rci *RedirectLocalCacheInstance) Set(key string, longURL string) {
	rci.mu.Lock()
	defer rci.mu.Unlock()
	if rci.maxSize == len(rci.data) {
		var keyToTheComingLRU = rci.data[rci.keyToLeastRecentlyUsed].previousEntry
		rci.data[keyToTheComingLRU].nextEntry = ""
		delete(rci.data, rci.keyToLeastRecentlyUsed)
		rci.keyToLeastRecentlyUsed = keyToTheComingLRU
	}
	// In both cases, a new entry is added.
	if len(rci.data) == 0 {
		rci.keyToLeastRecentlyUsed = key
	} else {
		rci.data[rci.keyToLastAdded].previousEntry = key
	}
	rci.data[key] = NewLocalCacheEntry(longURL)
	rci.data[key].nextEntry = rci.keyToLastAdded
	rci.keyToLastAdded = key
	rci.data[key].previousEntry = ""
}

func (rci *RedirectLocalCacheInstance) CleanUp() { // Receiver changed from 's' to 'rci'
	rci.mu.Lock()
	defer rci.mu.Unlock()
	deletedCount := 0 // Counter for the number of entries removed during this cleanup run.
	keysToDelete := []string{}
	var theoreticalSize = len(rci.data)
	for u, e := range rci.data {
		// Check if the current time is after the entry's hardTTL.
		// If true, the entry has expired.
		if time.Now().After(e.hardTTL) {
			if theoreticalSize > 1 {
				if rci.keyToLeastRecentlyUsed == u {
					rci.data[e.previousEntry].nextEntry = ""
					rci.keyToLeastRecentlyUsed = e.previousEntry
				} else if rci.keyToLastAdded == u {
					rci.data[e.nextEntry].previousEntry = ""
					rci.keyToLastAdded = e.nextEntry
				} else {
					rci.data[e.nextEntry].previousEntry = e.previousEntry
					rci.data[e.previousEntry].nextEntry = e.nextEntry
				}
				theoreticalSize--
			} else {
				rci.keyToLeastRecentlyUsed = ""
				rci.keyToLastAdded = ""
			}
			keysToDelete = append(keysToDelete, u)
		}
	}
	for _, u := range keysToDelete {
		delete(rci.data, u)
		deletedCount++
	}
	// Log a message after a cleanup.
	if deletedCount > 0 {
		log.Printf("Cache server: Removed %d expired entries during cleanup.", deletedCount)
	} else {
		log.Println("Cache server: No entries were removed")
	}
}

// Starts a new goroutine that periodically cleans up expired entries from the cache.
func (rci *RedirectLocalCacheInstance) startCleanupRoutine(interval time.Duration) {
	// Create a new ticker that will send a signal on its channel (ticker.C) every 'interval'.
	ticker := time.NewTicker(interval)
	// Ensures that the ticker is stopped when the function is terminated.
	defer ticker.Stop()
	// This is an infinite loop that listens for signals from the ticker's channel.
	// This ticker acts as a timer to trigger the periodic cleanup.
	for range ticker.C {
		rci.CleanUp()
	}
}

type RedisSimpleCache struct {
	client *redis.Client
	logger *zap.SugaredLogger
}

func NewRedisSimpleCache(client *redis.Client, logger *zap.SugaredLogger) *RedisSimpleCache {
	return &RedisSimpleCache{client: client, logger: logger}
}

func (rsc *RedisSimpleCache) Get(shortURL string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	val, err := rsc.client.Get(ctx, shortURL).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false
		}
		rsc.logger.Errorf("Error getting key '%s' from Redis: %v", shortURL, err)
		return "", false
	}
	return val, true
}

func (rsc *RedisSimpleCache) Set(shortURL string, longURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := rsc.client.Set(ctx, shortURL, longURL, 24*time.Hour).Err()
	if err != nil {
		rsc.logger.Errorf("Error setting key '%s' in Redis: %v", shortURL, err)
	}
}
