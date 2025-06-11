package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

const DEFAULT_PORT = "3000"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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
