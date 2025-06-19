package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/mactavishz/kuerzen/analytics/grpc"
	"github.com/mactavishz/kuerzen/shortener/lib"
	astore "github.com/mactavishz/kuerzen/store/analytics"
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
	client   *grpc.AnalyticsGRPCClient
}

func NewShortenHandler(urlStore store.URLStore, client *grpc.AnalyticsGRPCClient) *ShortenHandler {
	return &ShortenHandler{
		urlStore: urlStore,
		client:   client,
	}
}

func (h *ShortenHandler) HandleShortenURL(c *fiber.Ctx) error {
	req := new(ShortenURLRequest)
	evt := &astore.URLCreationEvent{
		ServiceName: "shortener",
		APIVer:      1,
		Success:     false,
		Timestamp:   time.Now(),
	}
	err := c.BodyParser(req)
	if err != nil {
		log.Printf("[HandleShortenURL] invalid request payload: %s\n", string(c.Body()))
		err = h.client.SendURLCreationEvent(context.TODO(), evt)
		if err != nil {
			log.Printf("[HandleShortenURL] failed to send event: %v\n", err)
		}
		return c.Status(fiber.StatusBadRequest).JSON(Envelope{
			"msg": "invalid request payload",
		})
	}
	evt.URL = req.URL
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(req)
	if err != nil {
		log.Printf("[HandleShortenURL] invalid url: %s\n", req.URL)
		err = h.client.SendURLCreationEvent(context.TODO(), evt)
		if err != nil {
			log.Printf("[HandleShortenURL] failed to send event: %v\n", err)
		}
		return c.Status(fiber.StatusBadRequest).JSON(Envelope{
			"msg": "invalid url",
		})
	}
	shortURL := lib.ToShortURL(req.URL, SHORT_URL_LENGTH)
	err = h.urlStore.CreateShortURL(shortURL, req.URL)
	if err != nil {
		evtErr := h.client.SendURLCreationEvent(context.TODO(), evt)
		if evtErr != nil {
			log.Printf("[HandleShortenURL] failed to send event: %v\n", evtErr)
		}
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
	evt.Success = true
	err = h.client.SendURLCreationEvent(context.TODO(), evt)
	if err != nil {
		log.Printf("[HandleShortenURL] failed to send event: %v\n", err)
	}
	return c.JSON(ShortenURLResponse{
		URL:     fmt.Sprintf("%s/%s", os.Getenv("KUERZEN_HOST"), shortURL),
		ShortID: shortURL,
	})
}
