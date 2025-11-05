package cache

import (
	"sync"
)

type LinkCache struct {
	mu    sync.RWMutex
	cache map[string]bool
}

func NewLinkCache() *LinkCache {
	return &LinkCache{
		cache: make(map[string]bool),
	}
}

func (c *LinkCache) Get(url string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, found := c.cache[url]
	return val, found
}

func (c *LinkCache) Set(url string, status bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[url] = status
}

//TODO : Implemet cache expiration time (e.g., 10 min TTL)
