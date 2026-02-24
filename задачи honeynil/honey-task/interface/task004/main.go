package main

// Задача: EventBus — подписка/отписка/публикация событий асинхронно (порядок сохраняется).

import (
	"errors"
	"fmt"
	"sync"
)

type Event struct {
	Type string
	Data interface{}
}

type EventHandler interface {
	Handle(event Event) error
}

type EventBus interface {
	Subscribe(eventType string, handler EventHandler) error
	Unsubscribe(eventType string, handler EventHandler) error
	Publish(event Event) error
	Close() error
}

type subscription struct {
	handler EventHandler
	ch      chan Event
}

type SimpleEventBus struct {
	mu   sync.RWMutex
	subs map[string][]*subscription
	wg   sync.WaitGroup
}

func NewSimpleEventBus() *SimpleEventBus {
	return &SimpleEventBus{subs: make(map[string][]*subscription)}
}

func (b *SimpleEventBus) Subscribe(eventType string, handler EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	sub := &subscription{handler: handler, ch: make(chan Event, 100)}
	b.subs[eventType] = append(b.subs[eventType], sub)
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for e := range sub.ch {
			sub.handler.Handle(e)
		}
	}()
	return nil
}

func (b *SimpleEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subs[eventType]
	for i, s := range subs {
		if s.handler == handler {
			close(s.ch)
			b.subs[eventType] = append(subs[:i], subs[i+1:]...)
			return nil
		}
	}
	return errors.New("handler not found")
}

func (b *SimpleEventBus) Publish(event Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, s := range b.subs[event.Type] {
		s.ch <- event
	}
	return nil
}

func (b *SimpleEventBus) Close() error {
	b.mu.Lock()
	for _, subs := range b.subs {
		for _, s := range subs {
			close(s.ch)
		}
	}
	b.subs = make(map[string][]*subscription)
	b.mu.Unlock()
	b.wg.Wait()
	return nil
}

// --- Демо ---

type logHandler struct{ name string }

func (h *logHandler) Handle(e Event) error {
	fmt.Printf("[%s] %s: %v\n", h.name, e.Type, e.Data)
	return nil
}

func main() {
	bus := NewSimpleEventBus()
	h1 := &logHandler{"handler1"}
	h2 := &logHandler{"handler2"}
	bus.Subscribe("user.created", h1)
	bus.Subscribe("user.created", h2)
	bus.Publish(Event{Type: "user.created", Data: "Alice"})
	bus.Publish(Event{Type: "user.created", Data: "Bob"})
	bus.Close()
}
