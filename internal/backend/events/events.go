package events

import (
	"context"
	"sync"
	"time"
)

type Event struct {
	Type      string
	ActorID   string
	ProjectID string
	ObjectID  string
	At        time.Time
	Data      map[string]any
}

type Handler func(context.Context, Event) error

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

func (b *Bus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *Bus) Publish(ctx context.Context, event Event) []error {
	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}

	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[event.Type]...)
	wildcard := append([]Handler(nil), b.handlers["*"]...)
	b.mu.RUnlock()

	handlers = append(handlers, wildcard...)

	var errs []error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
