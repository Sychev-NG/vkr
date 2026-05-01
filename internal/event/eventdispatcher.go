package event

import (
	"context"
	"log"
	"sync"
)

type Event interface {
	EventName() string
}

type EventHandler interface {
	Handle(ctx context.Context, event Event) error
}

type EventDispatcher struct {
	handlers 	map[string][]EventHandler
	mu			sync.RWMutex
}

func New() *EventDispatcher {
	return &EventDispatcher{handlers: map[string][]EventHandler{}}
}

func(d *EventDispatcher) Subscribe(handler EventHandler, event Event) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers[event.EventName()] = append(d.handlers[event.EventName()], handler)
}

func(d *EventDispatcher) Dispatch(ctx context.Context, event Event) {
	d.mu.RLock()
	handlers := d.handlers[event.EventName()]
	d.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			log.Printf("EventDispatcher::Dispatch [event:%s] Error: %v", event.EventName(), err)
		}
	}
}