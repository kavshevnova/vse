package main

// Задача: Observer — SimpleSubject, FilteredSubject, AsyncSubject.

import (
	"fmt"
	"sync"
)

type Observer interface {
	Update(event string, data interface{}) error
}

type Subject interface {
	Attach(observer Observer) error
	Detach(observer Observer) error
	Notify(event string, data interface{}) error
}

// --- SimpleSubject ---

type SimpleSubject struct {
	mu        sync.RWMutex
	observers []Observer
}

func (s *SimpleSubject) Attach(o Observer) error {
	s.mu.Lock()
	s.observers = append(s.observers, o)
	s.mu.Unlock()
	return nil
}

func (s *SimpleSubject) Detach(o Observer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, obs := range s.observers {
		if obs == o {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *SimpleSubject) Notify(event string, data interface{}) error {
	s.mu.RLock()
	obs := make([]Observer, len(s.observers))
	copy(obs, s.observers)
	s.mu.RUnlock()
	for _, o := range obs {
		if err := o.Update(event, data); err != nil {
			return err
		}
	}
	return nil
}

// --- FilteredSubject ---

type filteredEntry struct {
	observer Observer
	events   map[string]struct{}
}

type FilteredSubject struct {
	mu      sync.RWMutex
	entries []filteredEntry
}

func (f *FilteredSubject) AttachFiltered(o Observer, events ...string) error {
	m := make(map[string]struct{}, len(events))
	for _, e := range events {
		m[e] = struct{}{}
	}
	f.mu.Lock()
	f.entries = append(f.entries, filteredEntry{observer: o, events: m})
	f.mu.Unlock()
	return nil
}

func (f *FilteredSubject) Attach(o Observer) error { return f.AttachFiltered(o) }

func (f *FilteredSubject) Detach(o Observer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, e := range f.entries {
		if e.observer == o {
			f.entries = append(f.entries[:i], f.entries[i+1:]...)
			return nil
		}
	}
	return nil
}

func (f *FilteredSubject) Notify(event string, data interface{}) error {
	f.mu.RLock()
	entries := make([]filteredEntry, len(f.entries))
	copy(entries, f.entries)
	f.mu.RUnlock()
	for _, e := range entries {
		if len(e.events) == 0 {
			if err := e.observer.Update(event, data); err != nil {
				return err
			}
			continue
		}
		if _, ok := e.events[event]; ok {
			if err := e.observer.Update(event, data); err != nil {
				return err
			}
		}
	}
	return nil
}

// --- AsyncSubject ---

type asyncEntry struct {
	observer Observer
	ch       chan struct{ event string; data interface{} }
}

type AsyncSubject struct {
	mu      sync.RWMutex
	entries []*asyncEntry
}

func (a *AsyncSubject) Attach(o Observer) error {
	e := &asyncEntry{
		observer: o,
		ch:       make(chan struct{ event string; data interface{} }, 64),
	}
	a.mu.Lock()
	a.entries = append(a.entries, e)
	a.mu.Unlock()
	go func() {
		for msg := range e.ch {
			e.observer.Update(msg.event, msg.data)
		}
	}()
	return nil
}

func (a *AsyncSubject) Detach(o Observer) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, e := range a.entries {
		if e.observer == o {
			close(e.ch)
			a.entries = append(a.entries[:i], a.entries[i+1:]...)
			return nil
		}
	}
	return nil
}

func (a *AsyncSubject) Notify(event string, data interface{}) error {
	a.mu.RLock()
	entries := make([]*asyncEntry, len(a.entries))
	copy(entries, a.entries)
	a.mu.RUnlock()
	for _, e := range entries {
		e.ch <- struct{ event string; data interface{} }{event, data}
	}
	return nil
}

// --- Sample observers ---

type LogObserver struct{ Name string }

func (l *LogObserver) Update(event string, data interface{}) error {
	fmt.Printf("[%s] event=%s data=%v\n", l.Name, event, data)
	return nil
}

func main() {
	sub := &SimpleSubject{}
	o1 := &LogObserver{Name: "log1"}
	o2 := &LogObserver{Name: "log2"}
	sub.Attach(o1)
	sub.Attach(o2)
	sub.Notify("user.created", map[string]string{"id": "1", "name": "Alice"})
	sub.Detach(o1)
	sub.Notify("user.deleted", "1")

	fmt.Println("--- filtered ---")
	fs := &FilteredSubject{}
	fs.AttachFiltered(&LogObserver{Name: "filtered"}, "user.created")
	fs.Notify("user.created", "Bob")
	fs.Notify("user.deleted", "Bob") // не должен получить
}
