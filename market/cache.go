package market

import (
	"sync"
)

// KlineSharedPool shares market data to reduce memory usage
type KlineSharedPool struct {
	data     map[string]*Data // symbol -> full market data
	refCount map[string]int   // reference count
	mu       sync.RWMutex
}

// NewKlineSharedPool creates a new shared pool
func NewKlineSharedPool() *KlineSharedPool {
	return &KlineSharedPool{
		data:     make(map[string]*Data),
		refCount: make(map[string]int),
	}
}

// Get retrieves data from the pool
func (p *KlineSharedPool) Get(symbol string) *Data {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data[symbol]
}

// Put adds data to the pool
func (p *KlineSharedPool) Put(symbol string, data *Data) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data[symbol] = data
	p.refCount[symbol]++
}

// Retain increments reference count
func (p *KlineSharedPool) Retain(symbol string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.data[symbol]; exists {
		p.refCount[symbol]++
	}
}

// Release decrements reference count and removes if zero
func (p *KlineSharedPool) Release(symbol string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if count, exists := p.refCount[symbol]; exists {
		count--
		if count <= 0 {
			delete(p.data, symbol)
			delete(p.refCount, symbol)
		} else {
			p.refCount[symbol] = count
		}
	}
}

// Clear clears all data
func (p *KlineSharedPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data = make(map[string]*Data)
	p.refCount = make(map[string]int)
}
