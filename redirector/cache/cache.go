package cache

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type CacheProvider interface {
	Get(shortURL string) (string, bool)
	Set(shortURL string, longURL string)
}

const EXPIRATION_TIME = 10 * time.Minute

//const CACHE_CLEANUP_INTERVAL = 5 * time.Minute

type LocalCacheEntry struct {
	longURL string
	hardTTL time.Time
	prev    string // Key to the previous element
	next    string // Key to the next element.
}

//with prev and next the entries are concatenated with each other, thus simulating a list

func NewLocalCacheEntry(lURL string) *LocalCacheEntry {
	return &LocalCacheEntry{
		longURL: lURL,
		hardTTL: time.Now().Add(EXPIRATION_TIME),
		prev:    "",
		next:    "",
	}
}

// Maps are unordered, the concatenation (prev & next) of the map entries results in an order of the elements from last added to least recently used
type RedirectLocalCacheInstance struct {
	data    map[string]*LocalCacheEntry
	mu      sync.RWMutex
	maxSize int
	keyLA   string // Key to the last added entry of the local cache
	keyLRU  string // Key to the least recently used entry of the local cache
}

func NewRedirectLocalCacheInstance(maxSize int) (*RedirectLocalCacheInstance, error) {
	if maxSize <= 1 {
		return nil, errors.New("maxSize must be greater than 1 for RedirectLocalCacheInstance")
	}
	return &RedirectLocalCacheInstance{
		data:    make(map[string]*LocalCacheEntry),
		maxSize: maxSize,
	}, nil
}

func (rci *RedirectLocalCacheInstance) Get(shortURL string) (string, bool) {
	rci.mu.Lock()
	defer rci.mu.Unlock()
	entry, found := rci.data[shortURL]
	if !found {
		return "", false
	} else {
		// Check if entry has expired
		if time.Now().After(entry.hardTTL) {
			if len(rci.data) > 1 {
				if rci.keyLRU == shortURL {
					// Entry is the last one in the order
					// Predecessor of entry is the new LRU
					rci.data[entry.prev].next = ""
					rci.keyLRU = entry.prev
				} else if rci.keyLA == shortURL {
					// Entry is the first one in the order
					// Successor of entry is the new LA
					rci.data[entry.next].prev = ""
					rci.keyLA = entry.next
				} else {
					// Entry has predecessor and successor
					// Concatenate predecessor and successor
					rci.data[entry.next].prev = entry.prev
					rci.data[entry.prev].next = entry.next
				}
			} else {
				// Cache is empty
				rci.keyLRU = ""
				rci.keyLA = ""
			}
			// The real delete
			delete(rci.data, shortURL)
			return "", false
		}
	}
	// If there is a element for shortURL in the cache, the order of the entries must be changed
	// Eviction Policy: Least Recently Used
	// Cache: A <-> ... <-> D <-> E <-> F <-> ... <-> Z
	if entry.prev != "" {
		// Entry has a predecessor (prev is not an empty string), this predecessor should now get the position of the entry
		// The predecessor needs a new successor or becomes the new LRU
		rci.data[entry.prev].next = entry.next // A Key or an empty string
		if entry.next != "" {
			// Entry has a successor (next is not an empty string), this successor needs to know that it needs a new predecessor
			rci.data[entry.next].prev = entry.prev
		} else {
			// If next == "" (empty string), it means that the entry was the least recently used in the cache
			// This makes the predecessor the new LRU
			rci.keyLRU = entry.prev
		}
		// The last to be added is now the second last to be added and must know that he needs a predecessor
		rci.data[rci.keyLA].prev = shortURL
		// The entry becomes the new last added. No predecessor and successor is now the second last to be added
		entry.prev = ""
		entry.next = rci.keyLA
		//The cache must be informed of what was last added
		rci.keyLA = shortURL
	}
	//if prevEntry == "" it means that the entry is the first entry in the cache. No order to change
	entry.hardTTL = time.Now().Add(EXPIRATION_TIME)
	return entry.longURL, true
}

func (rci *RedirectLocalCacheInstance) Set(key string, longURL string) {
	rci.mu.Lock()
	defer rci.mu.Unlock()
	// Check whether entry should be updated
	entry, found := rci.data[key]
	if found {
		// Update the concatenation in the cache
		// If cache has only one or entry is already the LA -> order does not change
		if len(rci.data) > 1 || rci.keyLA == key {
			if rci.keyLRU == key {
				// Entry has no successor and its predecessor becomes the new LRU
				rci.data[entry.prev].next = ""
				rci.keyLRU = entry.prev
			}
			if rci.keyLA != key {
				// Entry has a successor and a predecessor. Concatenate successor and predecessor
				rci.data[entry.next].prev = entry.prev
				rci.data[entry.prev].next = entry.next
			}
			// Entry will be the new LA
			rci.data[key].prev = ""
			rci.data[rci.keyLA].prev = key
			rci.data[key].next = rci.keyLA
			rci.keyLA = key
		}
		// Update Entry with new longURL
		rci.data[key].longURL = longURL
		rci.data[key].hardTTL = time.Now().Add(EXPIRATION_TIME)
	} else {
		// Check to see if the capacity of the cache has been reached
		if rci.maxSize == len(rci.data) {
			// Get the second least recently used
			var keySLRU = rci.data[rci.keyLRU].prev
			// SLRU becomes the new LRU
			rci.data[keySLRU].next = ""
			// Delete last entry
			delete(rci.data, rci.keyLRU)
			// Define new LRU of the cache
			rci.keyLRU = keySLRU
		}
		// In both cases, a new entry is added.
		if len(rci.data) == 0 {
			rci.keyLRU = key
		} else {
			// Concatenate new element correctly
			rci.data[rci.keyLA].prev = key
		}
		rci.data[key] = NewLocalCacheEntry(longURL)
		rci.data[key].next = rci.keyLA
		rci.keyLA = key
		rci.data[key].prev = ""
	}
}

func (rci *RedirectLocalCacheInstance) CleanUp() { // Receiver changed from 's' to 'rci'
	rci.mu.Lock()
	defer rci.mu.Unlock()
	deletedCount := 0 // Counter for the number of entries removed during this cleanup run
	// Start at the LRU and work your way up to an entry where the TTL has not expired
	// Use the predecessor to follow the order in the simulated linked list
	// Find the first entry that has not yet expired, this is the new LRU
	// Up to this entry, all can be deleted
	maxIteration := len(rci.data)
	key := rci.keyLRU
	for i := 1; i <= maxIteration; i++ {
		if time.Now().After(rci.data[key].hardTTL) {
			if i == maxIteration {
				delete(rci.data, key)
				deletedCount++
				rci.keyLA = ""
				rci.keyLRU = ""
				break
			}
			key = rci.data[key].prev
			delete(rci.data, key)
			deletedCount++
		} else {
			rci.data[key].next = ""
			rci.keyLRU = key
			break
		}
	}
	// Log a message after a cleanup.
	if deletedCount > 0 {
		log.Printf("Cache server: Removed %d expired entries during cleanup.", deletedCount)
	} else {
		log.Println("Cache server: No entries were removed")
	}
}

// Starts a new goroutine that periodically cleans up expired entries from the cache.
func (rci *RedirectLocalCacheInstance) StartCleanupRoutine(interval time.Duration) {
	// Create a new ticker that will send a signal on its channel (ticker.C) every 'interval'.
	ticker := time.NewTicker(interval)
	// Ensures that the ticker is stopped when the function is terminated.
	defer ticker.Stop()
	// This is an infinite loop that listens for signals from the ticker's channel.
	// This ticker acts as a timer to trigger the periodic cleanup.
	for range ticker.C {
		rci.CleanUp()
	}
}

type RedisSimpleCache struct {
	client *redis.Client
	logger *zap.SugaredLogger
}

func NewRedisSimpleCache(client *redis.Client, logger *zap.SugaredLogger) *RedisSimpleCache {
	return &RedisSimpleCache{client: client, logger: logger}
}

func (rsc *RedisSimpleCache) Get(shortURL string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	val, err := rsc.client.Get(ctx, shortURL).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false
		}
		rsc.logger.Errorf("Error getting key '%s' from Redis: %v", shortURL, err)
		return "", false
	}
	return val, true
}

func (rsc *RedisSimpleCache) Set(shortURL string, longURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := rsc.client.Set(ctx, shortURL, longURL, 24*time.Hour).Err()
	if err != nil {
		rsc.logger.Errorf("Error setting key '%s' in Redis: %v", shortURL, err)
	}
}
