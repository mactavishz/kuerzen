package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/mactavishz/kuerzen/analytics/grpc"
	"github.com/mactavishz/kuerzen/redirector/api"
	"github.com/mactavishz/kuerzen/store/db"
	"github.com/mactavishz/kuerzen/store/migrations"
	store "github.com/mactavishz/kuerzen/store/url"
)

const DEFAULT_PORT = "3001"

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
	client, err := grpc.NewAnalyticsGRPCClient(os.Getenv("ANALYTICS_SERVICE_URL"))
	if err != nil {
		log.Fatalf("could not set up grpc client: %v", err)
	}
	handler := api.NewRedirectHandler(urlStore, client)

	app.Get("/api/v1/url/:shortURL", handler.HandleRedirect)

	port := os.Getenv("REDIRECTOR_PORT")
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	log.Printf("Redirector service listening on port :%s", port)
	app.Listen(":" + port)
}
