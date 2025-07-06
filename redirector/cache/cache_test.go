package cache

import (
	"testing"
	"time"
)

func TestNewRedirectLocalCacheInstance(t *testing.T) {

	_, err := NewRedirectLocalCacheInstance(0)
	if err == nil {
		t.Error("Expected error for maxSize <= 1, but got nil")
	}
	expectedErrorMsg := "maxSize must be greater than 1 for RedirectLocalCacheInstance"
	if err != nil && err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', but got '%v'", expectedErrorMsg, err)
	}

	_, err = NewRedirectLocalCacheInstance(1)
	if err == nil {
		t.Error("Expected error for maxSize <= 1, but got nil")
	}
	if err != nil && err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', but got '%v'", expectedErrorMsg, err)
	}

	cache, err := NewRedirectLocalCacheInstance(100)
	if err != nil {
		t.Errorf("Did not expect error for valid maxSize, but got: %v", err)
	}
	if cache == nil {
		t.Error("Expected cache instance, but got nil")
	}
	if cache.maxSize != 100 {
		t.Errorf("Expected maxSize to be 100, but got %d", cache.maxSize)
	}
	if len(cache.data) != 0 {
		t.Errorf("Expected empty cache data, but got %d entries", len(cache.data))
	}
	if cache.keyLA != "" || cache.keyLRU != "" {
		t.Errorf("Expected empty LRU/LA pointers, but got LA: %s, LRU: %s", cache.keyLA, cache.keyLRU)
	}
}

func TestSetNewEntry(t *testing.T) {
	cache, _ := NewRedirectLocalCacheInstance(3)

	cache.Set("short1", "long1")
	if len(cache.data) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(cache.data))
	}
	if cache.keyLA != "short1" || cache.keyLRU != "short1" {
		t.Errorf("Expected LA/LRU to be short1, got LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}
	val, ok := cache.data["short1"]
	if !ok || val.longURL != "long1" {
		t.Error("Entry short1 not found or value incorrect")
	}

	cache.Set("short2", "long2")
	if len(cache.data) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(cache.data))
	}
	if cache.keyLA != "short2" || cache.keyLRU != "short1" {
		t.Errorf("Expected LA:short2, LRU:short1, got LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}
	if cache.data["short2"].next != "short1" || cache.data["short1"].prev != "short2" {
		t.Error("Linked list for short1-short2 incorrect (between")
	}
	if cache.data["short2"].prev != "" || cache.data["short1"].next != "" {
		t.Error("Linked list for short1-short2 incorrect (start/end)")
	}

	cache.Set("short3", "long3")
	if len(cache.data) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(cache.data))
	}
	if cache.keyLA != "short3" || cache.keyLRU != "short1" {
		t.Errorf("Expected LA:short3, LRU:short1, got LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}
	if cache.data["short3"].next != "short2" || cache.data["short2"].prev != "short3" {
		t.Error("Linked list for short3-short2 incorrect")
	}
	if cache.data["short3"].prev != "" || cache.data["short1"].next != "" {
		t.Error("Linked list for short1-short2 incorrect (start/end)")
	}

	cache.Set("short4", "long4")
	if len(cache.data) != 3 {
		t.Errorf("Expected 3 entries after eviction, got %d", len(cache.data))
	}
	if _, ok := cache.data["short1"]; ok {
		t.Error("short1 was not removed")
	}
	if cache.keyLA != "short4" || cache.keyLRU != "short2" {
		t.Errorf("Expected LA:short4, LRU:short2 after eviction, got LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}
	if cache.data["short4"].next != "short3" || cache.data["short3"].prev != "short4" {
		t.Error("Linked list for short4-short3 incorrect")
	}
	if cache.data["short2"].next != "" {
		t.Error("short2 should be the end of the list")
	}
	if cache.data["short2"].prev != "short3" {
		t.Error("predecessor of short2 should be short3")
	}

	cache.Set("short3", "long3_updated")
	if len(cache.data) != 3 {
		t.Errorf("Expected 3 entries after update, got %d", len(cache.data))
	}
	if cache.keyLA != "short3" || cache.keyLRU != "short2" {
		t.Errorf("Expected LA:short3, LRU:short2 after update, got LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}
	if cache.data["short3"].longURL != "long3_updated" {
		t.Error("short3 value not updated")
	}
	if cache.data["short3"].prev != "" {
		t.Error("short3 should be the start of the list")
	}

	if cache.data["short3"].next != "short4" || cache.data["short4"].prev != "short3" {
		t.Error("Linked list short3-short4 after update incorrect")
	}
	if cache.data["short4"].next != "short2" || cache.data["short2"].prev != "short4" {
		t.Error("Linked list short4-short2 after update incorrect")
	}
	if cache.data["short2"].next != "" {
		t.Error("short2 should still be the LRU (end of list)")
	}
}

func TestGet(t *testing.T) {
	cache, _ := NewRedirectLocalCacheInstance(3)

	cache.Set("short1", "long1") // LA: short1, LRU: short1
	cache.Set("short2", "long2") // LA: short2, LRU: short1
	cache.Set("short3", "long3") // LA: short3, LRU: short1

	val, found := cache.Get("short3")
	if !found || val != "long3" {
		t.Error("Expected short3 to be found and value correct")
	}
	if cache.keyLA != "short3" {
		t.Errorf("Expected LA:short3 after Get, got %s", cache.keyLA)
	}
	if cache.keyLRU != "short1" {
		t.Errorf("Expected LRU:short1 after Get, got %s", cache.keyLRU)
	}
	if cache.data["short3"].prev != "" {
		t.Error("short3 should have no predecessor as it is LA")
	}
	if cache.data["short3"].next != "short2" || cache.data["short2"].prev != "short3" {
		t.Error("Linked list short3-short2 incorrect after Get")
	}
	if cache.data["short2"].next != "short1" || cache.data["short1"].prev != "short2" {
		t.Error("Linked list short2-short incorrect after Get")
	}
	if cache.data["short1"].next != "" {
		t.Error("short1 should be the end of the list after Get")
	}

	val, found = cache.Get("short1")
	if !found || val != "long1" {
		t.Error("Expected short1 to be found and value correct")
	}
	if cache.keyLA != "short1" {
		t.Errorf("Expected LA:short1 after Get, got %s", cache.keyLA)
	}
	if cache.keyLRU != "short2" {
		t.Errorf("Expected LRU:short2 after Get, got %s", cache.keyLRU)
	}
	if cache.data["short1"].prev != "" {
		t.Error("short1 should have no predecessor as it is LA")
	}
	if cache.data["short1"].next != "short3" || cache.data["short3"].prev != "short1" {
		t.Error("Linked list short1-short3 incorrect after Get")
	}
	if cache.data["short3"].next != "short2" || cache.data["short2"].prev != "short3" {
		t.Error("Linked list short3-short2 incorrect after Get")
	}
	if cache.data["short2"].next != "" {
		t.Error("short2 should be the end of the list after Get")
	}

	_, found = cache.Get("nonExistent")
	if found {
		t.Error("Did not expect 'nonExistent' to be found")
	}

	if cache.keyLA != "short1" || cache.keyLRU != "short2" {
		t.Errorf("Cache state changed after getting non-existent key, LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}

	cache.Set("expired", "expired_long")
	cache.data["expired"].hardTTL = time.Now().Add(-1 * time.Second)

	_, found = cache.Get("expired")
	if found {
		t.Error("Expected 'expired' to be deleted and not found")
	}
	if _, ok := cache.data["expired"]; ok {
		t.Error("Expired entry was not actually deleted from map")
	}

	if cache.keyLA != "short1" || cache.keyLRU != "short3" {
		t.Errorf("Pointers incorrect after deleting expired LA entry, LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}
	if cache.data["short1"].prev != "" {
		t.Error("short1 should have no predecessor as it is LA")
	}
	if cache.data["short1"].next != "short3" || cache.data["short3"].prev != "short1" {
		t.Error("Linked list short1-short3 incorrect after Get+Delete")
	}
	if cache.data["short3"].next != "" {
		t.Error("short3 should be the end of the list after Get+Delete")
	}
	if len(cache.data) != 2 {
		t.Errorf("Expected 2 entries after update, got %d", len(cache.data))
	}

	cache.Set("short4", "long4")
	val, found = cache.Get("short3")
	if !found || val != "long3" {
		t.Error("Expected short3 to be found and value correct")
	}
	if cache.keyLA != "short3" || cache.keyLRU != "short1" {
		t.Errorf("Pointers incorrect after deleting expired LA entry, LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}
	if cache.data["short3"].prev != "" {
		t.Error("short3 should have no predecessor as it is LA")
	}
	if cache.data["short3"].next != "short4" || cache.data["short4"].prev != "short3" {
		t.Error("Linked list short3-short4 incorrect after Get+Delete")
	}
	if cache.data["short4"].next != "short1" || cache.data["short1"].prev != "short4" {
		t.Error("Linked list short4-short1 incorrect after Get+Delete")
	}
	if cache.data["short1"].next != "" {
		t.Error("short1 should be the end of the list after Get+Delete")
	}
	if len(cache.data) != 3 {
		t.Errorf("Expected 3 entries after update, got %d", len(cache.data))
	}
}

func TestCleanUp(t *testing.T) {
	cache, _ := NewRedirectLocalCacheInstance(5)

	cache.Set("s1", "l1")
	cache.Set("s2", "l2")
	cache.Set("s3", "l3")
	cache.Set("s4", "l4")

	// Manipulate TTL
	cache.data["s1"].hardTTL = time.Now().Add(-2 * time.Second)
	cache.data["s2"].hardTTL = time.Now().Add(-1 * time.Second)
	// Happens automatically when adding
	// cache.data["s3"].hardTTL = time.Now().Add(10 * time.Minute)
	// cache.data["s4"].hardTTL = time.Now().Add(10 * time.Minute)

	cache.CleanUp()

	if len(cache.data) != 2 {
		t.Errorf("Expected 2 entries after cleanup, got %d", len(cache.data))
	}
	if _, ok := cache.data["s1"]; ok {
		t.Error("s1 (expired) was not removed by CleanUp")
	}
	if _, ok := cache.data["s2"]; ok {
		t.Error("s2 (expired) was not removed by CleanUp")
	}
	if _, ok := cache.data["s3"]; !ok {
		t.Error("s3 (not expired) was removed by CleanUp")
	}
	if _, ok := cache.data["s4"]; !ok {
		t.Error("s4 (not expired) was removed by CleanUp")
	}
	if cache.data["s4"].prev != "" {
		t.Error("s4 should have no predecessor as it is LA")
	}
	if cache.keyLA != "s4" || cache.keyLRU != "s3" {
		t.Errorf("Expected LA:s4, LRU:s3 after cleanup, got LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}
	if cache.data["s4"].next != "s3" || cache.data["s3"].prev != "s4" {
		t.Error("Linked list s4-s3 incorrect after cleanup")
	}
	if cache.data["s3"].next != "" {
		t.Error("s3 should be the new end of the list")
	}

	cache.data["s3"].hardTTL = time.Now().Add(-3 * time.Second)
	cache.data["s4"].hardTTL = time.Now().Add(-4 * time.Second)

	cache.CleanUp()
	if len(cache.data) != 0 {
		t.Errorf("Expected 0 entries after second cleanup, got %d", len(cache.data))
	}
	if cache.keyLA != "" || cache.keyLRU != "" {
		t.Errorf("Expected empty pointers after full cleanup, got LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}

	cache.CleanUp()
	if len(cache.data) != 0 {
		t.Errorf("Expected 0 entries after cleanup on empty cache, got %d", len(cache.data))
	}
}

func TestStartCleanupRoutine(t *testing.T) {
	cache, _ := NewRedirectLocalCacheInstance(5)
	cache.Set("s1", "l1")
	cache.Set("s2", "l2")
	cache.Set("s3", "l3")
	cache.Set("s4", "l4")
	cache.Set("s5", "l5")
	cache.data["s1"].hardTTL = time.Now().Add(25 * time.Millisecond)
	cache.data["s2"].hardTTL = time.Now().Add(55 * time.Millisecond)
	cache.data["s3"].hardTTL = time.Now().Add(57 * time.Millisecond)
	cache.data["s4"].hardTTL = time.Now().Add(115 * time.Millisecond)
	cache.data["s5"].hardTTL = time.Now().Add(145 * time.Millisecond)

	go cache.StartCleanupRoutine(30 * time.Millisecond)

	time.Sleep(31 * time.Millisecond)

	if _, found := cache.Get("s1"); found {
		t.Error("s1 should have been cleaned up by routine")
	}
	if len(cache.data) != 4 {
		t.Errorf("Expected 4 entry after routine, got %d", len(cache.data))
	}
	if cache.keyLA != "s5" || cache.keyLRU != "s2" {
		t.Errorf("Expected LA:s5, LRU:s2 after routine, got LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}
	if cache.data["s5"].prev != "" {
		t.Error("s5 should have no predecessor as it is LA")
	}
	if cache.data["s5"].next != "s4" || cache.data["s4"].prev != "s5" {
		t.Error("Linked list s5-s4 incorrect after routine")
	}
	if cache.data["s4"].next != "s3" || cache.data["s3"].prev != "s4" {
		t.Error("Linked list s4-s3 incorrect after routine")
	}
	if cache.data["s3"].next != "s2" || cache.data["s2"].prev != "s3" {
		t.Error("Linked list s3-s2 incorrect after routine")
	}
	if cache.data["s2"].next != "" {
		t.Error("s2 should be the new end of the list")
	}

	time.Sleep(30 * time.Millisecond)
	if _, found := cache.Get("s2"); found {
		t.Error("s2 should have been cleaned up by routine")
	}
	if _, found := cache.Get("s3"); found {
		t.Error("s3 should have been cleaned up by routine")
	}
	if len(cache.data) != 2 {
		t.Errorf("Expected 2 entry after routine, got %d", len(cache.data))
	}
	if cache.keyLA != "s5" || cache.keyLRU != "s4" {
		t.Errorf("Expected LA:s5, LRU:s4 after routine, got LA:%s, LRU:%s", cache.keyLA, cache.keyLRU)
	}
	if cache.data["s5"].prev != "" {
		t.Error("s5 should have no predecessor as it is LA")
	}
	if cache.data["s5"].next != "s4" || cache.data["s4"].prev != "s5" {
		t.Error("Linked list s5-s4 incorrect after routine")
	}
	if cache.data["s4"].next != "" {
		t.Error("s4 should be the new end of the list")
	}

	time.Sleep(30 * time.Millisecond)
	if len(cache.data) != 2 {
		t.Errorf("Expected 2 entry after routine, got %d", len(cache.data))
	}
	if _, ok := cache.data["s5"]; !ok {
		t.Error("s5 should still be present")
	}
	if _, ok := cache.data["s4"]; !ok {
		t.Error("s4 should still be present")
	}
	cache.Get("s4")
	cache.data["s4"].hardTTL = time.Now().Add(35 * time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	if len(cache.data) != 2 {
		t.Errorf("Expected 2 entry after routine, got %d", len(cache.data))
	}
	if _, ok := cache.data["s5"]; !ok {
		t.Error("s5 should still be present")
	}
	if _, ok := cache.data["s4"]; !ok {
		t.Error("s4 should still be present")
	}
	time.Sleep(30 * time.Millisecond)
	if len(cache.data) != 0 {
		t.Errorf("Expected 0 entry after routine, got %d", len(cache.data))
	}
}
