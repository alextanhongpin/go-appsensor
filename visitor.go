package main

import (
	"log"
	"sync"
	"time"
)

type Visitor struct {
	sync.RWMutex
	events         map[EventCode][]time.Time
	penalizedUntil time.Time
	count          int // Store the count of the number of times the visitor has been penalized. This rule can be part of the policy too.
	// isPenalized bool // This needs to be recomputed every time. Because the state needs to be reset once it has reached the penalized duration.
	// penalizedTime time.Time // Storing penalizedUntil saves you from storing the penalizedTime and penalized duration. To check whether the visitor is penalized is just a simple comparison of time.Before(penalizedUntil), or allowed is time.Now().After(penalizedUntil)
	// penalizedDuration time.Duration // Let the policy hold the policy duration. Then the visitor can hold the penalizedUntil. It is easier to check if the current time is before or after than the difference in time.
}

func NewVisitor() *Visitor {
	return &Visitor{
		// NOTE: Storing all the time will only grow the memory usage. Clear this up every interval.
		// But how to know which time to clear? Sort each policy by the Period, take the longest one,
		// And clear everything that is before time.Now().Add(-Period).
		events: make(map[EventCode][]time.Time),
	}
}

func (v *Visitor) ClearBefore(policies *PolicyManager) {
	v.Lock()
	for code, evts := range v.events {
		pol, _ := policies.Get(code)
		cp := evts[:0]
		for _, evt := range evts {
			if time.Since(evt) < pol.Period {
				log.Println("still valid timestamps")
				cp = append(cp, evt)
			}
		}
		v.events[code] = cp
	}
	v.Unlock()
}
func (v *Visitor) Add(code EventCode) {
	v.Lock()
	evts, ok := v.events[code]
	if !ok {
		v.events[code] = make([]time.Time, 0)
	}
	v.events[code] = append(evts, time.Now())
	v.Unlock()
}

func (v *Visitor) HasCount(code EventCode, elapsed time.Duration, threshold int) bool {
	v.RLock()
	evts, _ := v.events[code]
	v.RUnlock()
	var i int
	// t := time.Now().Add(-elapsed)
	for _, evt := range evts {
		// if t.After(evt) {
		if time.Since(evt) < elapsed {
			i++
			if i >= threshold {
				return true
			}
		}
	}
	return false
}
func (v *Visitor) Penalize(duration time.Duration) {
	v.Lock()
	// Stack the penalty. If there's an existing penalty...
	// NOTE: It's not possible to stack if the user is already blocked when they are penalized.
	if v.penalizedUntil.After(time.Now()) {
		log.Println("stacked", v.penalizedUntil)
		v.penalizedUntil = v.penalizedUntil.Add(duration)
	} else {
		log.Println("first time", v.penalizedUntil)
		v.penalizedUntil = time.Now().Add(duration)
	}
	v.Unlock()
}
