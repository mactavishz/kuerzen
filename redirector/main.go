package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

const DEFAULT_PORT = "3001"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := fiber.New()

	app.Get("/api/v1/url/:shortURL", func(c *fiber.Ctx) error {
		return c.SendString("TODO: " + c.Params("shortURL"))
	})

	port := os.Getenv("REDIRECTOR_PORT")
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	log.Printf("Redirector service listening on port :%s", port)
	app.Listen(":" + port)
}
