package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
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

	app := fiber.New()
	urlStore := store.NewPostgresURLStore(pgDB)
	handler := api.NewShortenHandler(urlStore)

	app.Post("/api/v1/url/shorten", handler.HandleShortenURL)

	port := os.Getenv("SHORTENER_PORT")
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	log.Printf("Shortener service listening on port :%s", port)
	app.Listen(":" + port)
}
