package appsensor

import (
	"sync"
	"time"
)

var (
	REQ1 = "REQ1"
	REQ2 = "REQ2"
	REQ3 = "REQ3"
)

type (
	AppSensor interface {
		Allow(id string) bool
		Log(id, evt string)
		Start()
		Stop()
	}

	Event struct {
		Code      string
		Count     int
		CreatedAt time.Time
		ID        string
		Penalized bool
		UpdatedAt time.Time
	}
)

func NewEvent(id, code string) *Event {
	return &Event{
		Code:      code,
		Count:     1,
		CreatedAt: time.Now().UTC(),
		ID:        id,
		UpdatedAt: time.Now().UTC(),
	}
}

type appSensorImpl struct {
	sync.RWMutex
	cache map[string]*Event

	ch        chan *Event
	duration  time.Duration
	quit      chan struct{}
	threshold int
}

func (a *appSensorImpl) Start() {
	go a.loop()
}

func (a *appSensorImpl) loop() {
	for {
		select {
		case <-a.quit:
			return
		case evt, ok := <-a.ch:
			if !ok {
				return
			}
			
			if c, found := a.get(evt.ID); found {
				c.Count++
				if a.block(c) {
					a.unblock(c)
				}
			} else {
				// Create a new event.
				a.set(evt.ID, evt)
			}
		}
	}
}

func (a *appSensorImpl) Stop() {
	close(a.quit)
}

func (a *appSensorImpl) Log(id, evt string) {
	select {
	case <-a.quit:
		return
	case a.ch <- NewEvent(id, evt):
	}
}

func (a *appSensorImpl) Allow(id string) bool {
	evt, found := a.get(id)
	if found && a.block(evt) {
		return a.unblock(evt)
	}
	return !found
}

func (a *appSensorImpl) get(id string) (*Event, bool){
	a.RLock()
	evt, found := a.cache[id]
	a.RUnlock()
	return evt, found
}

func (a *appSensorImpl) set(id string, evt *Event) {
	a.Lock()
	a.cache[id] = evt
	a.Unlock()
} 

func (a *appSensorImpl) unblock(evt *Event) bool {
	if evt.Penalized && time.Since(evt.UpdatedAt) > a.duration {
		evt.Count = 0
		evt.Penalized = false
	}
	return !evt.Penalized
}

func (a *appSensorImpl) block(evt *Event) bool {
	if !evt.Penalized && evt.Count >= a.threshold {
		evt.Penalized = true
		evt.UpdatedAt = time.Now().UTC()
	}
	return evt.Penalized
}
