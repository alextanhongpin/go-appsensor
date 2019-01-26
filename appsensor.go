package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	InvalidPassword = "invalid_password"
)

var Events = [...]string{InvalidPassword}

type Rule interface {
	Validate(store EventStore, evt Event) bool
	Update(store EventStore, evt Event)
}

var invalidPasswordRule *InvalidPasswordRule
var defaultRule *DefaultRule

func init() {
	invalidPasswordRule = &InvalidPasswordRule{3, 1 * time.Second}
	defaultRule = &DefaultRule{1, 500 * time.Millisecond}
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
	duration  time.Duration
}

func (d *DefaultRule) Validate(store EventStore, evt Event) bool {
	e := store.Get(evt)
	return e.Count >= d.threshold
}

func update(store EventStore, evt Event, duration time.Duration) {
	meta := store.Get(evt)
	if time.Since(meta.UpdatedAt) > (time.Duration(meta.Count) * duration) {
		meta.Count--
		fmt.Println("decrement evt", evt, meta.Count)
		if meta.Count < 0 {
			store.Delete(evt)
			fmt.Println("cleared evt", evt)
			return
		}
		meta.UpdatedAt = time.Now().UTC()
		store.Put(evt, meta)
	}
}

func (d *DefaultRule) Update(store EventStore, evt Event) {
	update(store, evt, d.duration)
}

type InvalidPasswordRule struct {
	threshold int
	duration  time.Duration
}

func (i *InvalidPasswordRule) Validate(store EventStore, evt Event) bool {
	e := store.Get(evt)
	return e.Count > i.threshold
}

func (i *InvalidPasswordRule) Update(store EventStore, evt Event) {
	update(store, evt, i.duration)
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
	Allow(id string) bool
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
	size := e.store.Size()
	events := e.store.Keys(max(20/100*size, size))
	for _, evt := range events {
		rule := RuleStrategy(evt)
		rule.Update(e.store, evt)
	}
}

func (e *EventManagerImpl) Log(evt Event) {
	meta := e.store.Get(evt)

	e.mu.Lock()
	meta.Count++
	meta.UpdatedAt = time.Now().UTC()
	e.store.Put(evt, meta)
	e.mu.Unlock()
	fmt.Println("increment key", evt.ID, meta.Count)
}

func (e *EventManagerImpl) Allow(id string) bool {
	// For the given user id/ip, if any of the rule is broken,
	// block them.
	for _, evtType := range Events {
		evt := Event{id, evtType}
		rule := RuleStrategy(evt)
		return !rule.Validate(e.store, evt)
	}
	return true
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
		for i := 0; i < 5; i++ {
			evtMgr.Log(evt)
		}
		fmt.Println(evtMgr.Allow(evt.ID))
	}

	go func() {
		for {
			time.Sleep(10 * time.Second)
			n := rand.Intn(5)
			for _, evt := range events {
				for i := 0; i < n; i++ {
					evtMgr.Log(evt)
				}
			}
		}
	}()

	go func() {
		// Periodically check if the event can be called.
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				for _, evt := range events {
					if evtMgr.Allow(evt.ID) {
						fmt.Println(evt.ID, evt.Type, "is unblocked")
					} else {
						fmt.Println(evt.ID, evt.Type, "is blocked")
					}
				}
			}
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
