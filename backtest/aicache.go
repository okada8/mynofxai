package backtest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"nofx/kernel"
	"nofx/market"
)

type cachedDecision struct {
	Key           string               `json:"key"`
	PromptVariant string               `json:"prompt_variant"`
	Timestamp     int64                `json:"ts"`
	Decision      *kernel.FullDecision `json:"decision"`
}

// AICache persists AI decisions for repeated backtesting or replay.
type AICache struct {
	mu      sync.RWMutex
	path    string
	Entries map[string]cachedDecision `json:"entries"`
}

func LoadAICache(path string) (*AICache, error) {
	if path == "" {
		return nil, fmt.Errorf("ai cache path is empty")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	cache := &AICache{
		path:    path,
		Entries: make(map[string]cachedDecision),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return cache, nil
	}
	if err := json.Unmarshal(data, cache); err != nil {
		return nil, err
	}
	if cache.Entries == nil {
		cache.Entries = make(map[string]cachedDecision)
	}
	return cache, nil
}

func (c *AICache) Path() string {
	if c == nil {
		return ""
	}
	return c.path
}

func (c *AICache) Get(key string) (*kernel.FullDecision, bool) {
	if c == nil || key == "" {
		return nil, false
	}
	c.mu.RLock()
	entry, ok := c.Entries[key]
	c.mu.RUnlock()
	if !ok || entry.Decision == nil {
		return nil, false
	}
	return cloneDecision(entry.Decision), true
}

func (c *AICache) Put(key string, variant string, ts int64, decision *kernel.FullDecision) error {
	if c == nil || key == "" || decision == nil {
		return nil
	}
	entry := cachedDecision{
		Key:           key,
		PromptVariant: variant,
		Timestamp:     ts,
		Decision:      cloneDecision(decision),
	}
	c.mu.Lock()
	c.Entries[key] = entry
	c.mu.Unlock()
	return c.save()
}

func (c *AICache) save() error {
	if c == nil || c.path == "" {
		return nil
	}
	c.mu.RLock()
	data, err := json.MarshalIndent(c, "", "  ")
	c.mu.RUnlock()
	if err != nil {
		return err
	}
	return writeFileAtomic(c.path, data, 0o644)
}

func cloneDecision(src *kernel.FullDecision) *kernel.FullDecision {
	if src == nil {
		return nil
	}
	data, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	var dst kernel.FullDecision
	if err := json.Unmarshal(data, &dst); err != nil {
		return nil
	}
	return &dst
}

func computeCacheKey(ctx *kernel.Context, variant string, ts int64) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("context is nil")
	}
	payload := struct {
		Variant        string                 `json:"variant"`
		Timestamp      int64                  `json:"ts"`
		CurrentTime    string                 `json:"current_time"`
		Account        kernel.AccountInfo     `json:"account"`
		Positions      []kernel.PositionInfo  `json:"positions"`
		CandidateCoins []kernel.CandidateCoin `json:"candidate_coins"`
		MarketData     map[string]market.Data `json:"market"`
		MarginUsedPct  float64                `json:"margin_used_pct"`
		Runtime        int                    `json:"runtime_minutes"`
		CallCount      int                    `json:"call_count"`
	}{
		Variant:        variant,
		Timestamp:      ts,
		CurrentTime:    ctx.CurrentTime,
		Account:        ctx.Account,
		Positions:      ctx.Positions,
		CandidateCoins: ctx.CandidateCoins,
		MarginUsedPct:  ctx.Account.MarginUsedPct,
		Runtime:        ctx.RuntimeMinutes,
		CallCount:      ctx.CallCount,
		MarketData:     make(map[string]market.Data, len(ctx.MarketDataMap)),
	}

	for symbol, data := range ctx.MarketDataMap {
		if data == nil {
			continue
		}
		payload.MarketData[symbol] = *data
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(bytes)
	return hex.EncodeToString(sum[:]), nil
}

// ----------------------------------------------------------------------------
// Generic LLM Cache (Scheme A, B, C)
// ----------------------------------------------------------------------------

type LLMCacheEntry struct {
	Response  string `json:"response"`
	Timestamp int64  `json:"ts"`
	Model     string `json:"model,omitempty"`
}

// LLMCache implements a generic tiered cache for LLM responses.
// It supports memory (L1) and disk (L2) storage, and prompt deduplication.
type LLMCache struct {
	mu       sync.RWMutex
	memCache map[string]LLMCacheEntry // L1: Memory cache
	diskDir  string                   // L2: Disk cache directory
	ttl      time.Duration            // Time to live for cache entries
}

var (
	globalLLMCache     *LLMCache
	globalLLMCacheOnce sync.Once
)

// GetGlobalLLMCache returns the singleton LLM cache instance.
func GetGlobalLLMCache() *LLMCache {
	globalLLMCacheOnce.Do(func() {
		// Default location: user's home directory or temp
		homeDir, _ := os.UserHomeDir()
		cacheDir := filepath.Join(homeDir, ".nofi", "llm_cache")
		globalLLMCache, _ = NewLLMCache(cacheDir, 24*time.Hour)
	})
	return globalLLMCache
}

// NewLLMCache creates a new generic LLM cache.
func NewLLMCache(diskDir string, ttl time.Duration) (*LLMCache, error) {
	if diskDir != "" {
		if err := os.MkdirAll(diskDir, 0o700); err != nil {
			return nil, err
		}
	}
	return &LLMCache{
		memCache: make(map[string]LLMCacheEntry),
		diskDir:  diskDir,
		ttl:      ttl,
	}, nil
}

// Get retrieves a cached response for the given prompt and model.
func (c *LLMCache) Get(prompt string, model string) (string, bool) {
	if c == nil {
		return "", false
	}

	key := c.computeKey(prompt, model)

	// 1. Check Memory Cache (L1)
	c.mu.RLock()
	entry, ok := c.memCache[key]
	c.mu.RUnlock()

	if ok {
		if c.isExpired(entry.Timestamp) {
			c.mu.Lock()
			delete(c.memCache, key)
			c.mu.Unlock()
			return "", false
		}
		return entry.Response, true
	}

	// 2. Check Disk Cache (L2)
	if c.diskDir != "" {
		entry, ok := c.readFromDisk(key)
		if ok {
			if c.isExpired(entry.Timestamp) {
				return "", false
			}
			// Promote to memory
			c.mu.Lock()
			c.memCache[key] = entry
			c.mu.Unlock()
			return entry.Response, true
		}
	}

	return "", false
}

// Put stores a response in the cache.
func (c *LLMCache) Put(prompt string, model string, response string) error {
	if c == nil {
		return nil
	}

	key := c.computeKey(prompt, model)
	entry := LLMCacheEntry{
		Response:  response,
		Timestamp: time.Now().UnixMilli(),
		Model:     model,
	}

	// 1. Write to Memory (L1)
	c.mu.Lock()
	c.memCache[key] = entry
	c.mu.Unlock()

	// 2. Write to Disk (L2)
	if c.diskDir != "" {
		return c.writeToDisk(key, entry)
	}
	return nil
}

// computeKey generates a SHA256 hash of the prompt and model.
// Implements Scheme C (Smart Deduplication) partially by normalizing prompt if needed.
func (c *LLMCache) computeKey(prompt string, model string) string {
	// Normalize prompt: remove leading/trailing whitespace
	normalizedPrompt := strings.TrimSpace(prompt)
	
	// We can add more sophisticated normalization here if needed (e.g. regex to remove timestamps)
	// For now, exact match on trimmed prompt is safest.

	input := fmt.Sprintf("%s|%s", model, normalizedPrompt)
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

func (c *LLMCache) isExpired(ts int64) bool {
	if c.ttl <= 0 {
		return false
	}
	return time.Since(time.UnixMilli(ts)) > c.ttl
}

func (c *LLMCache) readFromDisk(key string) (LLMCacheEntry, bool) {
	path := filepath.Join(c.diskDir, key+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return LLMCacheEntry{}, false
	}

	var entry LLMCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return LLMCacheEntry{}, false
	}
	return entry, true
}

func (c *LLMCache) writeToDisk(key string, entry LLMCacheEntry) error {
	path := filepath.Join(c.diskDir, key+".json")
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return writeFileAtomic(path, data, 0o644)
}
