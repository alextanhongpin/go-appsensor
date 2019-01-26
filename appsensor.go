package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	InvalidPassword = "invalid_password"
)

type Rule interface {
	Validate(store EventStore, evt Event) bool
}

var invalidPasswordRule *InvalidPasswordRule
var defaultRule *DefaultRule

func init() {
	invalidPasswordRule = &InvalidPasswordRule{3}
	defaultRule = &DefaultRule{1}
}

func RuleStrategy(evt Event) Rule {
	switch evt.Type {
	case InvalidPassword:
		// Use prototype pattern.
		rule := *invalidPasswordRule
		return &rule
	default:
		rule := *defaultRule
		return &rule
	}
}

type DefaultRule struct {
	threshold int
}

func (d *DefaultRule) Validate(store EventStore, evt Event) bool {
	e := store.Get(evt)
	return e.Count > d.threshold
}

type InvalidPasswordRule struct {
	threshold int
}

func (i *InvalidPasswordRule) Validate(store EventStore, evt Event) bool {
	e := store.Get(evt)
	return e.Count > i.threshold
}

type Event struct {
	ID   string
	Type string
}

type EventMetadata struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Count     int
	Frozen    bool
}

type EventStore interface {
	Get(evt Event) EventMetadata
	Put(evt Event, meta EventMetadata)
	Delete(evt Event)
	Size() int
	Keys(n int) []Event
}

type EventStoreImpl struct {
	mu     *sync.RWMutex
	events map[Event]EventMetadata
}

func NewEventStore() *EventStoreImpl {
	return &EventStoreImpl{
		mu:     new(sync.RWMutex),
		events: make(map[Event]EventMetadata),
	}
}

func (e *EventStoreImpl) Get(evt Event) EventMetadata {
	e.mu.RLock()
	ee := e.events[evt]
	e.mu.RUnlock()
	return ee
}

func (e *EventStoreImpl) Put(evt Event, meta EventMetadata) {
	e.mu.Lock()
	e.events[evt] = meta
	e.mu.Unlock()
}

func (e *EventStoreImpl) Delete(evt Event) {
	e.mu.Lock()
	delete(e.events, evt)
	e.mu.Unlock()
}

func (e *EventStoreImpl) Size() int {
	e.mu.RLock()
	count := len(e.events)
	e.mu.RUnlock()
	return count
}

func (e *EventStoreImpl) Keys(n int) []Event {
	e.mu.Lock()
	defer e.mu.Unlock()
	var keys []Event
	var i int
	for key := range e.events {
		keys = append(keys, key)
		i++
		if i >= n {
			break
		}
	}
	return keys
}

type EventManager interface {
	Log(Event)
	Allow(Event) bool
}

type EventManagerImpl struct {
	quit  chan interface{}
	mu    *sync.RWMutex
	store EventStore
}

func NewEventManager() *EventManagerImpl {
	return &EventManagerImpl{
		quit:  make(chan interface{}),
		mu:    new(sync.RWMutex),
		store: NewEventStore(),
	}
}

func (e *EventManagerImpl) loop() {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-e.quit:
			return
		case <-t.C:
			if size := e.store.Size(); size > 0 {
				e.cleanup()
			}
		}
	}
}

func (e *EventManagerImpl) cleanup() {
	lockDuration := 1 * time.Second
	size := e.store.Size()
	keys := e.store.Keys(max(20/100*size, size))
	for _, key := range keys {
		meta := e.store.Get(key)
		// The lock duration is proportional to the frequency of the event.
		if time.Since(meta.UpdatedAt) > (time.Duration(meta.Count) * lockDuration) {
			meta.Count--
			fmt.Println("decrement key", key, meta.Count)
			if meta.Count < 0 {
				e.store.Delete(key)
				fmt.Println("cleared key", key)
				continue
			}
			meta.UpdatedAt = time.Now().UTC()
			e.store.Put(key, meta)
		}
	}
}

func (e *EventManagerImpl) Log(evt Event) {
	meta := e.store.Get(evt)

	e.mu.Lock()
	meta.Count++
	meta.UpdatedAt = time.Now().UTC()
	e.store.Put(evt, meta)
	e.mu.Unlock()
	fmt.Println("increment key", meta.Count)
}

func (e *EventManagerImpl) Allow(evt Event) bool {
	rule := RuleStrategy(evt)
	return !rule.Validate(e.store, evt)
}

func main() {
	events := []Event{
		Event{"1", "invalid_password"},
		Event{"2", "invalid_password"},
		Event{"3", "something else"},
	}

	evtMgr := NewEventManager()

	var wg sync.WaitGroup
	wg.Add(1)
	go evtMgr.loop()
	go func() {
		time.Sleep(60 * time.Second)
		wg.Done()
	}()
	for _, evt := range events {
		for i := 0; i < 3; i++ {
			evtMgr.Log(evt)
		}
		fmt.Println(evtMgr.Allow(evt))
	}

	go func() {
		time.Sleep(5 * time.Second)
		for _, evt := range events {
			evtMgr.Log(evt)
		}
	}()

	wg.Wait()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
