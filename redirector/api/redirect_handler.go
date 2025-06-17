package api

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	store "github.com/mactavishz/kuerzen/store/url"
)

type Envelope map[string]any

type RedirectHandler struct {
	urlStore store.URLStore
}

func NewRedirectHandler(urlStore store.URLStore) *RedirectHandler {
	return &RedirectHandler{
		urlStore: urlStore,
	}
}

func (h *RedirectHandler) HandleRedirect(c *fiber.Ctx) error {
	shortURL := c.Params("shortURL")
	// TODO: handle caching of redirects
	longURL, err := h.urlStore.GetLongURL(shortURL)
	if err != nil {
		if errors.Is(err, store.ErrShortURLNotFound) {
			log.Printf("[HandleRedirect] short URL not found: %s\n", shortURL)
			return c.Status(fiber.StatusNotFound).SendString("Not Found")
		}
		log.Printf("[HandleRedirect] failed to get long URL: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	// use 307 to prevent browsers from caching the redirect
	return c.Redirect(longURL, fiber.StatusTemporaryRedirect)
}
