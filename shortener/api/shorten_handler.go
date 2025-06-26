package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/mactavishz/kuerzen/analytics/grpc"
	"github.com/mactavishz/kuerzen/shortener/lib"
	astore "github.com/mactavishz/kuerzen/store/analytics"
	store "github.com/mactavishz/kuerzen/store/url"
	"go.uber.org/zap"
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
	client   *grpc.AnalyticsGRPCClient
	logger   *zap.SugaredLogger
}

func NewShortenHandler(urlStore store.URLStore, client *grpc.AnalyticsGRPCClient, logger *zap.SugaredLogger) *ShortenHandler {
	return &ShortenHandler{
		urlStore: urlStore,
		client:   client,
		logger:   logger,
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
		h.logger.Infow("invalid request payload", "payload", string(c.Body()))
		err = h.client.SendURLCreationEvent(context.TODO(), evt)
		if err != nil {
			h.logger.Errorf("failed to send event: %v\n", err)
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "invalid request payload",
		})
	}
	evt.URL = req.URL
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(req)
	if err != nil {
		h.logger.Infow("invalid url", "url", req.URL)
		err = h.client.SendURLCreationEvent(context.TODO(), evt)
		if err != nil {
			h.logger.Errorf("failed to send event: %v\n", err)
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "invalid url",
		})
	}
	shortURL := lib.ToShortURL(req.URL, SHORT_URL_LENGTH)
	err = h.urlStore.CreateShortURL(shortURL, req.URL)
	if err != nil {
		evtErr := h.client.SendURLCreationEvent(context.TODO(), evt)
		if evtErr != nil {
			h.logger.Errorf("failed to send event: %v\n", evtErr)
		}
		if errors.Is(err, store.ErrDuplicateLongURL) {
			h.logger.Infow("long URL already exists", "url", req.URL)
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"msg": "long URL already exists",
			})
		} else {
			h.logger.Errorf("failed to create short URL: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "failed to create short URL",
			})
		}
	}
	evt.Success = true
	err = h.client.SendURLCreationEvent(context.TODO(), evt)
	if err != nil {
		h.logger.Errorf("failed to send event: %v\n", err)
	}
	h.logger.Infow("short URL created", "shortURL", shortURL, "longURL", req.URL)
	return c.JSON(ShortenURLResponse{
		URL:     fmt.Sprintf("%s/%s", os.Getenv("KUERZEN_HOST"), shortURL),
		ShortID: shortURL,
	})
}
