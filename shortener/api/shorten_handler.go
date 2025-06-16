package api

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/mactavishz/kuerzen/shortener/lib"
	store "github.com/mactavishz/kuerzen/store/url"
)

const SHORT_URL_LENGTH = 8

type ShortenURLRequest struct {
	URL string `json:"url" validate:"required,http_url,max=1024"`
}

type ShortenURLResponse struct {
	URL     string `json:"url"`
	ShortID string `json:"short_id"`
}

type ShortenHandler struct {
	urlStore store.URLStore
}

func NewShortenHandler(urlStore store.URLStore) *ShortenHandler {
	return &ShortenHandler{
		urlStore: urlStore,
	}
}

func (h *ShortenHandler) HandleShortenURL(c *fiber.Ctx) error {
	req := new(ShortenURLRequest)
	err := c.BodyParser(req)
	if err != nil {
		return err
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(req)
	if err != nil {
		return err
	}
	shortURL := lib.ToShortURL(req.URL, SHORT_URL_LENGTH)
	err = h.urlStore.CreateShortURL(shortURL)
	if err != nil {
		return err
	}
	return c.JSON(ShortenURLResponse{
		URL:     fmt.Sprintf("%s/%s", os.Getenv("KUERZEN_HOST"), shortURL),
		ShortID: shortURL,
	})
}
