package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/mactavishz/kuerzen/shortener/migrations"
	"github.com/mactavishz/kuerzen/shortener/store"
)

const DEFAULT_PORT = "3000"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	pgDB, err := store.Open()
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	err = store.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		log.Fatalf("could not run database migrations: %v", err)
	}

	app := fiber.New()

	app.Post("/api/v1/url/shorten", func(c *fiber.Ctx) error {
		return c.SendString("TODO!")
	})

	port := os.Getenv("SHORTENER_PORT")
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	log.Printf("Shortener service listening on port :%s", port)
	app.Listen(":" + port)
}
