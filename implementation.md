## New Version

```go
package main

import (
	"log"
	"sync"
	"time"
)

type ThreatCode string

type ThreatBitwise uint

const (
	BadPassword ThreatCode = "bad_password"
)

type Threat struct {
	Code      ThreatCode
	Threshold int
	Period    time.Duration
}

func (t *Threat) Allow(visitor *Visitor) bool {
	var i int
	for _, evt := range visitor.events {
		if evt.code != t.Code {
			continue
		}
		if time.Since(evt.createdAt) < t.Period {
			i++
			if i > t.Threshold {
				return false
			}
		}
	}
	return true
}

type ThreatManager struct {
	sync.RWMutex
	threats map[ThreatCode]*Threat
}

func NewThreatManager(threats map[ThreatCode]*Threat) *ThreatManager {
	return &ThreatManager{
		threats: threats,
	}
}

func (t *ThreatManager) Get(code ThreatCode) (*Threat, bool) {
	t.Lock()
	th, ok := t.threats[code]
	t.Unlock()
	return th, ok
}

type Event struct {
	code      ThreatCode
	createdAt time.Time
}

type Visitor struct {
	sync.RWMutex
	ID ID
	// something to check if the code already exists...codes []ThreatCode
	// Bitwise operator?
	events []Event
}

func (v *Visitor) Add(evt Event) {
	v.Lock()
	v.events = append([]Event{evt}, v.events...)
	v.Unlock()
}

type ID string

type VisitorManager struct {
	sync.RWMutex
	visitors map[ID]*Visitor
}

func NewVisitorManager() *VisitorManager {
	return &VisitorManager{
		visitors: make(map[ID]*Visitor),
	}
}

func (v *VisitorManager) Add(id ID) {
	v.Lock()
	v.visitors[id] = &Visitor{ID: id, events: make([]Event, 0)}
	v.Unlock()
}

func (v *VisitorManager) Get(id ID) (*Visitor, bool) {
	v.RLock()
	vis, ok := v.visitors[id]
	v.RUnlock()
	return vis, ok
}

func NewBadPasswordEvent() Event {
	return Event{code: BadPassword, createdAt: time.Now()}
}
func main() {
	threatMgr := NewThreatManager(map[ThreatCode]*Threat{
		BadPassword: &Threat{
			Code:      BadPassword,
			Threshold: 3,
			Period:    1 * time.Second,
		},
	})

	threat, _ := threatMgr.Get(BadPassword)

	visitorMgr := NewVisitorManager()
	id := ID("0.0.0.0")
	visitorMgr.Add(id)
	visitor, _ := visitorMgr.Get(id)
	visitor.Add(NewBadPasswordEvent())
	visitor.Add(NewBadPasswordEvent())
	visitor.Add(NewBadPasswordEvent())
	visitor.Add(NewBadPasswordEvent())
	{
		time.Sleep(1 * time.Second)
		visitor, _ := visitorMgr.Get(id)
		log.Println(threat.Allow(visitor))
	}
}

// At every endpoint, get the user unique id (client ip)
// Loop through each events
// For each event with a threat, increment the counter
// For each threat, compare the threshold counter versus the existing counter
// If the threshold exceed, break the operation and return false
```
