package backtest

import (
	"sync"
	"time"
)

type cacheEntry struct {
	values    []float64
	timestamp time.Time
}

// IndicatorCache caches technical indicator values to avoid redundant calculations
type IndicatorCache struct {
	mu    sync.RWMutex
	cache map[string]cacheEntry // key -> entry
	pool  sync.Pool             // Reusable []float64 slices
	ttl   time.Duration         // Time to live for cache entries
}

// NewIndicatorCache creates a new indicator cache with default TTL
func NewIndicatorCache() *IndicatorCache {
	return &IndicatorCache{
		cache: make(map[string]cacheEntry),
		pool: sync.Pool{
			New: func() interface{} {
				// Default capacity, will grow if needed
				return make([]float64, 0, 1000)
			},
		},
		ttl: 5 * time.Minute, // Default TTL
	}
}

// SetTTL sets the expiration duration for cache entries
func (c *IndicatorCache) SetTTL(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttl = duration
}

// Get retrieves cached values
func (c *IndicatorCache) Get(key string) ([]float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	entry, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	// Check TTL
	if time.Since(entry.timestamp) > c.ttl {
		// Lazy expiration: we don't delete here to avoid upgrading lock, 
		// just return miss. Cleanup can happen on Put or dedicated cleanup.
		return nil, false
	}

	return entry.values, true
}

// Put stores values in cache. It makes a copy of values to avoid race conditions.
func (c *IndicatorCache) Put(key string, values []float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Use pool for storage
	dest := c.pool.Get().([]float64)
	if cap(dest) < len(values) {
		dest = make([]float64, len(values))
	} else {
		dest = dest[:len(values)]
	}
	copy(dest, values)
	
	c.cache[key] = cacheEntry{
		values:    dest,
		timestamp: time.Now(),
	}
}

// Clear clears the cache
func (c *IndicatorCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Reset map
	c.cache = make(map[string]cacheEntry)
}
