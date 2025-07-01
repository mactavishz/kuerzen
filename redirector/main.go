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
	combinedCache "github.com/mactavishz/kuerzen/redirector/cache"
	database "github.com/mactavishz/kuerzen/store/db"
	"github.com/mactavishz/kuerzen/store/migrations"
	store "github.com/mactavishz/kuerzen/store/url"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const DEFAULT_PORT = "3001"

func main() {
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
