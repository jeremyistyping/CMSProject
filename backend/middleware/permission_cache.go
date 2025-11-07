package middleware

import (
	"fmt"
	"sync"
	"time"
)

// PermissionCacheEntry stores cached permission result
type PermissionCacheEntry struct {
	HasPermission bool
	ExpiresAt     time.Time
}

// PermissionCache provides simple in-memory caching for permissions
type PermissionCache struct {
	cache map[string]PermissionCacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewPermissionCache creates a new permission cache with specified TTL
func NewPermissionCache(ttl time.Duration) *PermissionCache {
	cache := &PermissionCache{
		cache: make(map[string]PermissionCacheEntry),
		ttl:   ttl,
	}
	
	// Start cleanup goroutine
	go cache.cleanupExpired()
	
	return cache
}

// Get retrieves cached permission result
func (pc *PermissionCache) Get(userID uint, module, action string) (bool, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	
	key := pc.cacheKey(userID, module, action)
	entry, exists := pc.cache[key]
	
	if !exists || time.Now().After(entry.ExpiresAt) {
		return false, false
	}
	
	return entry.HasPermission, true
}

// Set stores permission result in cache
func (pc *PermissionCache) Set(userID uint, module, action string, hasPermission bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	
	key := pc.cacheKey(userID, module, action)
	pc.cache[key] = PermissionCacheEntry{
		HasPermission: hasPermission,
		ExpiresAt:     time.Now().Add(pc.ttl),
	}
}

// InvalidateUser clears all cached permissions for a user
func (pc *PermissionCache) InvalidateUser(userID uint) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	
	prefix := fmt.Sprintf("user:%d:", userID)
	for key := range pc.cache {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			delete(pc.cache, key)
		}
	}
}

// Clear removes all cached permissions
func (pc *PermissionCache) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	
	pc.cache = make(map[string]PermissionCacheEntry)
}

// cacheKey generates unique cache key
func (pc *PermissionCache) cacheKey(userID uint, module, action string) string {
	return fmt.Sprintf("user:%d:module:%s:action:%s", userID, module, action)
}

// cleanupExpired removes expired entries periodically
func (pc *PermissionCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		pc.mu.Lock()
		now := time.Now()
		for key, entry := range pc.cache {
			if now.After(entry.ExpiresAt) {
				delete(pc.cache, key)
			}
		}
		pc.mu.Unlock()
	}
}
