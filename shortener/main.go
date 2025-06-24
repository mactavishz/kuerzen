package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	grpc "github.com/mactavishz/kuerzen/analytics/grpc"
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

	urlStore := store.NewPostgresURLStore(pgDB)
	client, err := grpc.NewAnalyticsGRPCClient(os.Getenv("ANALYTICS_SERVICE_URL"))
	if err != nil {
		log.Fatalf("could not set up grpc client: %v", err)
	}
	handler := api.NewShortenHandler(urlStore, client)

	app.Post("/api/v1/url/shorten", timeout.NewWithContext(handler.HandleShortenURL, 3*time.Second))
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "service": "shortener"})
	})

	port := os.Getenv("SHORTENER_PORT")
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	log.Printf("Shortener service listening on port :%s", port)
	app.Listen(":" + port)
}
