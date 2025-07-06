package api

import (
	"context"
	"errors"
	"github.com/mactavishz/kuerzen/retries"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mactavishz/kuerzen/analytics/grpc"
	astore "github.com/mactavishz/kuerzen/store/analytics"
	store "github.com/mactavishz/kuerzen/store/url"
	"go.uber.org/zap"
)

type RedirectHandler struct {
	urlStore store.URLStore
	client   *grpc.AnalyticsGRPCClient
	logger   *zap.SugaredLogger
}

func NewRedirectHandler(urlStore store.URLStore, client *grpc.AnalyticsGRPCClient, logger *zap.SugaredLogger) *RedirectHandler {
	return &RedirectHandler{
		urlStore: urlStore,
		client:   client,
		logger:   logger,
	}
}

func (h *RedirectHandler) HandleRedirect(c *fiber.Ctx) error {
	shortURL := c.Params("shortURL")
	// TODO: handle caching of redirects
	rfo := retries.RetryWithExponentialBackoff(h.urlStore.GetLongURL(shortURL, c.Context()))
	longURL, ok := rfo.Rest[0].(string)
	if !ok {
		h.logger.Errorf("Failed to cast longURL from rfo.Rest[0] to string. Received type: %T", rfo.Rest[0])
		return errors.New("invalid type for long URL in retry result")
	}
	err := rfo.Err
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
			h.logger.Infow("short URL not found", "shortURL", shortURL)
			err = retries.RetryWithExponentialBackoff(h.client.SendURLRedirectEvent(context.TODO(), evt)).Err
			if err != nil {
				h.logger.Errorf("failed to send event: %v\n", err)
			}
			return c.Status(fiber.StatusNotFound).SendString("Not Found")
		}
		err = retries.RetryWithExponentialBackoff(h.client.SendURLRedirectEvent(context.TODO(), evt)).Err
		if err != nil {
			h.logger.Errorf("failed to send event: %v\n", err)
		}
		h.logger.Errorf("failed to get long URL: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	evt.Success = true
	err = retries.RetryWithExponentialBackoff(h.client.SendURLRedirectEvent(context.TODO(), evt)).Err
	if err != nil {
		h.logger.Errorf("failed to send event: %v\n", err)
	}
	// use 307 to prevent browsers from caching the redirect
	h.logger.Infow("request redirected", "shortURL", shortURL, "longURL", longURL)
	return c.Redirect(longURL, fiber.StatusTemporaryRedirect)
}
