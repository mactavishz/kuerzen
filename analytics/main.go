package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

const DEFAULT_PORT = "3002"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := fiber.New()

	app.Post("/api/v1/events", func(c *fiber.Ctx) error {
		return c.SendString("TODO")
	})

	port := os.Getenv("ANALYTICS_PORT")
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	log.Printf("Analytics service listening on port :%s", port)
	app.Listen(":" + port)
}
