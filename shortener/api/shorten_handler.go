package api

import (
	"github.com/gofiber/fiber/v2"
	store "github.com/mactavishz/kuerzen/store/url"
)

type ShortenHandler struct {
	urlStore store.URLStore
}

func NewShortenHandler(urlStore store.URLStore) *ShortenHandler {
	return &ShortenHandler{
		urlStore: urlStore,
	}
}

func (h *ShortenHandler) HandleShortenURL(c *fiber.Ctx) error {
	panic("TODO")
}
