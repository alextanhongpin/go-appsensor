package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	InvalidPassword = "invalid_password"
	Default         = "default"
)

type Rule interface {
	Allow(store EventStore, evt Event) bool
	Update(store EventStore, evt Event)
	Respond(store EventStore, evt Event)
	Type() string
}

type RuleManager interface {
	Strategy(Event) Rule
	Types() []string
}

func NewRuleManager(rules ...Rule) *RuleManagerImpl {
	ruleMgr := &RuleManagerImpl{
		rules: make(map[string]Rule),
		// By default, block the client when the
		// event threshold hits 1 for a block duration of 1 hour.
		defaults: &DefaultRule{
			threshold: 1,
			duration:  1 * time.Hour,
			ruleType:  Default,
		},
	}
	for _, rule := range rules {
		ruleMgr.rules[rule.Type()] = rule
	}
	return ruleMgr
}

type RuleManagerImpl struct {
	rules    map[string]Rule
	defaults Rule
}

func (r *RuleManagerImpl) Types() []string {
	// Can be cached to reduce read.
	var types []string
	for key := range r.rules {
		types = append(types, key)
	}
	return types
}

func (r *RuleManagerImpl) Strategy(evt Event) Rule {
	rule, exist := r.rules[evt.Type]
	if !exist {
		return r.defaults
	}
	return rule
}

type DefaultRule struct {
	threshold int
	ruleType  string
	duration  time.Duration
}

func (d *DefaultRule) Type() string {
	return d.ruleType
}

func (d *DefaultRule) Allow(store EventStore, evt Event) bool {
	e := store.Get(evt)
	return e.Count < d.threshold
	// return e.IsFrozen
}

func update(store EventStore, evt Event, duration time.Duration) {
	meta := store.Get(evt)
	if time.Since(meta.UpdatedAt) > (time.Duration(meta.Count) * duration) {
		meta.Count--
		// meta.Frozen = meta.Count >= threshold
		log.Printf("[%s] Reducing penalty for user %s, threshold is now %d\n", evt.Type, evt.ID, meta.Count)
		if meta.Count < 0 {
			store.Delete(evt)
			log.Printf("[%s] user %s is unblocked\n", evt.Type, evt.ID)
			return
		}
		meta.UpdatedAt = time.Now().UTC()
		store.Put(evt, meta)
	}
}

func (d *DefaultRule) Update(store EventStore, evt Event) {
	update(store, evt, d.duration)
}

func (d *DefaultRule) Respond(store EventStore, evt Event) {
	meta := store.Get(evt)
	meta.Count++
	meta.UpdatedAt = time.Now().UTC()

	log.Printf("[%s] Increasing penalty for user %s, threshold is now %d\n", evt.Type, evt.ID, meta.Count)

	store.Put(evt, meta)
}

type InvalidPasswordRule struct {
	DefaultRule
	warnThreshold     int
	criticalThreshold int
}

func (i *InvalidPasswordRule) Allow(store EventStore, evt Event) bool {
	return i.DefaultRule.Allow(store, evt)
}

func (i *InvalidPasswordRule) Update(store EventStore, evt Event) {
	update(store, evt, i.duration)
}

func (i *InvalidPasswordRule) Respond(store EventStore, evt Event) {
	i.DefaultRule.Respond(store, evt)
	meta := store.Get(evt)
	if meta.Count >= i.criticalThreshold {
		log.Printf("[%s] Threshold exceeded for user %s by %d\n", evt.Type, evt.ID, meta.Count)
		return
	}
	if meta.Count >= i.warnThreshold {
		log.Printf("[%s] Warn threshold exceeded for user %s by %d\n", evt.Type, evt.ID, meta.Count)
	}
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
	Clear(id string)
}

type EventManagerImpl struct {
	quit    chan interface{}
	store   EventStore
	ruleMgr RuleManager
	cache   *UserCache
}

type UserCache struct {
	mu *sync.RWMutex
	// Cache to check if the id exist.
	cache map[string]struct{}
}

func NewUserCache() *UserCache {
	return &UserCache{
		mu:    new(sync.RWMutex),
		cache: make(map[string]struct{}),
	}
}

func (u *UserCache) Exist(id string) bool {
	u.mu.RLock()
	_, exist := u.cache[id]
	u.mu.RUnlock()
	return exist
}

func (u *UserCache) Add(id string) {
	u.mu.Lock()
	u.cache[id] = struct{}{}
	u.mu.Unlock()
}

func (u *UserCache) Delete(id string) {
	u.mu.Lock()
	delete(u.cache, id)
	u.mu.Unlock()
}

func NewEventManager(rules ...Rule) *EventManagerImpl {
	return &EventManagerImpl{
		quit:    make(chan interface{}),
		store:   NewEventStore(),
		ruleMgr: NewRuleManager(rules...),
		cache:   NewUserCache(),
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
		rule := e.ruleMgr.Strategy(evt)
		rule.Update(e.store, evt)
	}
}

func (e *EventManagerImpl) Log(evt Event) {
	rule := e.ruleMgr.Strategy(evt)
	rule.Respond(e.store, evt)
	e.cache.Add(evt.ID)
}

func (e *EventManagerImpl) Allow(id string) bool {
	// For the given user id/ip, if any of the rule is broken,
	// block them.
	// If the user does not exist, just allow them.
	if exist := e.cache.Exist(id); !exist {
		return true
	}
	for _, evtType := range e.ruleMgr.Types() {
		evt := Event{id, evtType}
		rule := e.ruleMgr.Strategy(evt)
		if isAllowed := rule.Allow(e.store, evt); !isAllowed {
			return false
		}
	}
	return true
}

func (e *EventManagerImpl) Clear(id string) {
	for _, evtType := range e.ruleMgr.Types() {
		evt := Event{id, evtType}
		e.store.Delete(evt)
	}
	e.cache.Delete(id)
}

func main() {
	rules := []Rule{
		&InvalidPasswordRule{
			DefaultRule: DefaultRule{
				threshold: 3,
				duration:  1 * time.Second,
				ruleType:  InvalidPassword,
			},
			warnThreshold:     5,
			criticalThreshold: 15,
		},
		&DefaultRule{
			threshold: 1,
			duration:  1 * time.Second,
			ruleType:  Default,
		},
	}

	events := []Event{
		Event{"1", InvalidPassword},
		Event{"2", InvalidPassword},
		Event{"3", Default},
		Event{"3", InvalidPassword},
	}

	evtMgr := NewEventManager(rules...)

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
	}

	go func() {
		for {
			time.Sleep(15 * time.Second)
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
						log.Printf("[%s] user %s is unblocked", evt.Type, evt.ID)
					} else {
						log.Printf("[%s] user %s is blocked", evt.Type, evt.ID)
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
