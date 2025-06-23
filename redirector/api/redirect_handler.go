package api

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mactavishz/kuerzen/analytics/grpc"
	astore "github.com/mactavishz/kuerzen/store/analytics"
	store "github.com/mactavishz/kuerzen/store/url"
)

type RedirectHandler struct {
	urlStore store.URLStore
	client   *grpc.AnalyticsGRPCClient
}

func NewRedirectHandler(urlStore store.URLStore, client *grpc.AnalyticsGRPCClient) *RedirectHandler {
	return &RedirectHandler{
		urlStore: urlStore,
		client:   client,
	}
}

func (h *RedirectHandler) HandleRedirect(c *fiber.Ctx) error {
	shortURL := c.Params("shortURL")
	// TODO: handle caching of redirects
	longURL, err := h.urlStore.GetLongURL(shortURL)
	evt := &astore.URLRedirectEvent{
		ServiceName: "redirector",
		APIVer:      1,
		Success:     false,
		ShortURL:    shortURL,
		LongURL:     longURL,
		Timestamp:   time.Now(),
	}
	if err != nil {
		if errors.Is(err, store.ErrShortURLNotFound) {
			log.Printf("[HandleRedirect] short URL not found: %s\n", shortURL)
			err = h.client.SendURLRedirectEvent(context.TODO(), evt)
			if err != nil {
				log.Printf("[HandleRedirect] failed to send event: %v\n", err)
			}
			return c.Status(fiber.StatusNotFound).SendString("Not Found")
		}
		err = h.client.SendURLRedirectEvent(context.TODO(), evt)
		if err != nil {
			log.Printf("[HandleRedirect] failed to send event: %v\n", err)
		}
		log.Printf("[HandleRedirect] failed to get long URL: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	evt.Success = true
	err = h.client.SendURLRedirectEvent(context.TODO(), evt)
	if err != nil {
		log.Printf("[HandleRedirect] failed to send event: %v\n", err)
	}
	// use 307 to prevent browsers from caching the redirect
	return c.Redirect(longURL, fiber.StatusTemporaryRedirect)
}
