package api

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/mactavishz/kuerzen/shortener/lib"
	store "github.com/mactavishz/kuerzen/store/url"
)

const SHORT_URL_LENGTH = 8

type Envelope map[string]any

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
		log.Printf("[HandleShortenURL] invalid request payload: %s\n", string(c.Body()))
		return c.Status(fiber.StatusBadRequest).JSON(Envelope{
			"msg": "invalid request payload",
		})
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(req)
	if err != nil {
		log.Printf("[HandleShortenURL] invalid url: %s\n", req.URL)
		return c.Status(fiber.StatusBadRequest).JSON(Envelope{
			"msg": "invalid url",
		})
	}
	shortURL := lib.ToShortURL(req.URL, SHORT_URL_LENGTH)
	err = h.urlStore.CreateShortURL(shortURL, req.URL)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateLongURL) {
			log.Printf("[HandleShortenURL] long URL already exists: %s\n", req.URL)
			return c.Status(fiber.StatusConflict).JSON(Envelope{
				"msg": "long URL already exists",
			})
		} else {
			log.Printf("[HandleShortenURL] failed to create short URL: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(Envelope{
				"msg": "failed to create short URL",
			})
		}
	}
	return c.JSON(ShortenURLResponse{
		URL:     fmt.Sprintf("%s/%s", os.Getenv("KUERZEN_HOST"), shortURL),
		ShortID: shortURL,
	})
}
