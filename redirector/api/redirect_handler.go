package api

import (
	"context"
	"errors"
	"github.com/mactavishz/kuerzen/retries"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mactavishz/kuerzen/analytics/grpc"
	"github.com/mactavishz/kuerzen/redirector/cache"
	astore "github.com/mactavishz/kuerzen/store/analytics"
	store "github.com/mactavishz/kuerzen/store/url"
	"go.uber.org/zap"
)

type RedirectHandler struct {
	urlStore      store.URLStore
	client        *grpc.AnalyticsGRPCClient
	logger        *zap.SugaredLogger
	localCache    cache.CacheProvider
	externalCache cache.CacheProvider
}

func NewRedirectHandler(urlStore store.URLStore, client *grpc.AnalyticsGRPCClient, logger *zap.SugaredLogger, localCache cache.CacheProvider, externalCache cache.CacheProvider) *RedirectHandler {
	return &RedirectHandler{
		urlStore:      urlStore,
		client:        client,
		logger:        logger,
		localCache:    localCache,
		externalCache: externalCache,
	}
}

func (h *RedirectHandler) HandleRedirect(c *fiber.Ctx) error {
	shortURL := c.Params("shortURL")
	var longURL string
	var found bool
	evt := &astore.URLRedirectEvent{
		ServiceName: "redirector",
		APIVer:      1,
		Success:     false,
		ShortURL:    shortURL,
		LongURL:     longURL,
		Timestamp:   time.Now(),
	}

	longURL, found = h.localCache.Get(shortURL)
	if found {
		h.logger.Infof("Cache Hit: Local Cache for shortURL: %s", shortURL)
		return h.performRedirect(c, evt, shortURL, longURL)
	}
	h.logger.Infof("Cache Miss: Local Cache for shortURL: %s", shortURL)

	longURL, found = h.externalCache.Get(shortURL)
	if found {
		h.logger.Infof("Cache Hit: External Cache for shortURL: %s", shortURL)
		h.localCache.Set(shortURL, longURL)
		return h.performRedirect(c, evt, shortURL, longURL)
	}
	h.logger.Infof("Cache Miss: External Cache for shortURL: %s", shortURL)

	rfo := retries.RetryWithExponentialBackoff(h.urlStore.GetLongURL(shortURL, c.Context()))
	longURL, ok := rfo.Rest[0].(string)
	if !ok {
		h.logger.Errorf("Failed to cast longURL from rfo.Rest[0] to string. Received type: %T", rfo.Rest[0])
		return errors.New("invalid type for long URL in retry result")
	}
	err := rfo.Err
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

	h.logger.Infof("DB Hit: Found %s in DB. Populating caches.", shortURL)
	h.localCache.Set(shortURL, longURL)
	h.externalCache.Set(shortURL, longURL)

	evt.Success = true
	err = retries.RetryWithExponentialBackoff(h.client.SendURLRedirectEvent(context.TODO(), evt)).Err
	if err != nil {
		h.logger.Errorf("failed to send event for %s: %v\n", shortURL, err)
	}
	// use 307 to prevent browsers from caching the redirect
	h.logger.Infow("request redirected", "shortURL", shortURL, "longURL", longURL)
	return c.Redirect(longURL, fiber.StatusTemporaryRedirect)
}
