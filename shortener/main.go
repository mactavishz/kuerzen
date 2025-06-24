package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	grpc "github.com/mactavishz/kuerzen/analytics/grpc"
	"github.com/mactavishz/kuerzen/middleware/loadshed"
	"github.com/mactavishz/kuerzen/shortener/api"
	"github.com/mactavishz/kuerzen/store/db"
	"github.com/mactavishz/kuerzen/store/migrations"
	store "github.com/mactavishz/kuerzen/store/url"
)

const DEFAULT_PORT = "3000"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	pgDB, err := db.Open()
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	err = db.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		log.Fatalf("could not run database migrations: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:   "shortener",
		BodyLimit: 1024 * 1024 * 1, // 1MB
	})

	loadshedMiddleware, err := loadshed.NewLoadSheddingMiddleware(loadshed.Config{
		CPUThreshold: 0.9,
		MemThreshold: 0.9,
		Interval:     500 * time.Millisecond,
	})
	if err != nil {
		log.Fatalf("could not set up load shedding middleware: %v", err)
	}
	app.Use(loadshedMiddleware)

	urlStore := store.NewPostgresURLStore(pgDB)
	client, err := grpc.NewAnalyticsGRPCClient(os.Getenv("ANALYTICS_SERVICE_URL"))
	if err != nil {
		log.Fatalf("could not set up grpc client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing grpc client: %v", err)
		}
	}()
	defer func() {
		if err := pgDB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()
	handler := api.NewShortenHandler(urlStore, client)

	app.Post("/api/v1/url/shorten", timeout.NewWithContext(handler.HandleShortenURL, 3*time.Second))
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "service": "shortener"})
	})

	port := os.Getenv("SHORTENER_PORT")
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	// Listen from a different goroutine
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Panic(err)
		}
	}()
	log.Printf("Shortener service listening on port :%s", port)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("Gracefully shutting down...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Error shutting down: %v", err)
	}
	log.Println("Server was gracefully shut down")
}
