package agent

import (
	"sync"
)

// EventType defines the type of event
type EventType string

const (
	EventMarketUpdate    EventType = "market_update"
	EventAlphaSignal     EventType = "alpha_signal"
	EventRiskAlert       EventType = "risk_alert"
	EventTradeExecution  EventType = "trade_execution"
)

// Event represents a system event
type Event struct {
	Type      EventType   `json:"type"`
	Source    string      `json:"source"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

// EventHandler handles events
type EventHandler func(event Event)

// EventBus manages event subscriptions and publishing
type EventBus struct {
	subscribers map[EventType][]EventHandler
	mu          sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]EventHandler),
	}
}

// Subscribe subscribes a handler to an event type
func (b *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

// Publish publishes an event to all subscribers
func (b *EventBus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	if handlers, ok := b.subscribers[event.Type]; ok {
		for _, handler := range handlers {
			// Execute handlers asynchronously to avoid blocking
			go handler(event)
		}
	}
}
