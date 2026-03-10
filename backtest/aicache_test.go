package backtest

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLLMCache_Memory(t *testing.T) {
	cache, err := NewLLMCache("", 1*time.Minute) // No disk cache
	assert.NoError(t, err)

	prompt := "What is Bitcoin?"
	model := "gpt-4"
	response := "Bitcoin is a decentralized digital currency."

	// Test Put
	err = cache.Put(prompt, model, response)
	assert.NoError(t, err)

	// Test Get
	cachedResp, ok := cache.Get(prompt, model)
	assert.True(t, ok)
	assert.Equal(t, response, cachedResp)

	// Test Miss
	_, ok = cache.Get("Unknown prompt", model)
	assert.False(t, ok)
}

func TestLLMCache_DiskPersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "llm_cache_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	cache1, err := NewLLMCache(tempDir, 1*time.Minute)
	assert.NoError(t, err)

	prompt := "Explain Quantum Computing"
	model := "claude-3"
	response := "Quantum computing uses quantum bits..."

	// Put in cache1
	err = cache1.Put(prompt, model, response)
	assert.NoError(t, err)

	// Verify in cache1
	val, ok := cache1.Get(prompt, model)
	assert.True(t, ok)
	assert.Equal(t, response, val)

	// Create new cache instance pointing to same directory (simulate restart)
	cache2, err := NewLLMCache(tempDir, 1*time.Minute)
	assert.NoError(t, err)

	// Verify in cache2 (should load from disk)
	val2, ok := cache2.Get(prompt, model)
	assert.True(t, ok)
	assert.Equal(t, response, val2)
}

func TestLLMCache_TTL(t *testing.T) {
	cache, err := NewLLMCache("", 100*time.Millisecond) // Short TTL
	assert.NoError(t, err)

	prompt := "Quick prompt"
	model := "fast-model"
	response := "Quick response"

	cache.Put(prompt, model, response)

	// Immediate check
	val, ok := cache.Get(prompt, model)
	assert.True(t, ok)
	assert.Equal(t, response, val)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Check again
	_, ok = cache.Get(prompt, model)
	assert.False(t, ok)
}

func TestLLMCache_Deduplication(t *testing.T) {
	cache, err := NewLLMCache("", 1*time.Minute)
	assert.NoError(t, err)

	model := "gpt-3.5"
	response := "42"

	// Prompt with extra spaces
	prompt1 := "  What is the answer?  "
	prompt2 := "What is the answer?"

	cache.Put(prompt1, model, response)

	// Should match prompt2 due to normalization
	val, ok := cache.Get(prompt2, model)
	assert.True(t, ok)
	assert.Equal(t, response, val)
}

func TestGlobalLLMCache(t *testing.T) {
	cache := GetGlobalLLMCache()
	assert.NotNil(t, cache)
	
	// Ensure singleton
	cache2 := GetGlobalLLMCache()
	assert.Equal(t, cache, cache2)
}
