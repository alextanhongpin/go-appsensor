```go
package main

import (
	"log"
	"sync"
	"time"
)

type EventCode uint

type Visitor struct {
	sync.RWMutex
	events         map[EventCode][]time.Time
	penalizedUntil time.Time
	// isPenalized bool // This needs to be recomputed every time. Because the state needs to be reset once it has reached the penalized duration.
	// penalizedTime time.Time // Storing penalizedUntil saves you from storing the penalizedTime and penalized duration. To check whether the visitor is penalized is just a simple comparison of time.Before(penalizedUntil), or allowed is time.Now().After(penalizedUntil)
	// penalizedDuration time.Duration // Let the policy hold the policy duration. Then the visitor can hold the penalizedUntil. It is easier to check if the current time is before or after than the difference in time.
}

func NewVisitor() *Visitor {
	return &Visitor{
		events: make(map[EventCode][]time.Time),
	}
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
	if v.penalizedUntil.After(time.Now()) {
		log.Println("stacked", v.penalizedUntil)
		v.penalizedUntil = v.penalizedUntil.Add(duration)
	} else {
		log.Println("first time", v.penalizedUntil)
		v.penalizedUntil = time.Now().Add(duration)
	}
	v.Unlock()
}

const (
	BadPassword EventCode = iota
	BadRequest
)

type Policy struct {
	Code              EventCode
	Threshold         int
	Period            time.Duration // Period is the time range where the threshold is honoured.
	PenalizedDuration time.Duration
}
type PolicyManager struct {
	sync.RWMutex
	policies []Policy
	plain    Policy
}

func NewPolicyManager() *PolicyManager {
	return &PolicyManager{
		policies: make([]Policy, 0),
		plain:    Policy{Threshold: 3, PenalizedDuration: 10 * time.Minute},
	}
}
func (p *PolicyManager) Add(pol ...Policy) {
	p.Lock()
	p.policies = append(p.policies, pol...)
	p.Unlock()
}

func (p *PolicyManager) Get(code EventCode) (Policy, bool) {
	p.RLock()
	for _, pol := range p.policies {
		if pol.Code == code {
			p.RUnlock()
			return pol, true
		}
	}
	p.RUnlock()
	return p.plain, false
}

type VisitorManager struct {
	// Slice would be faster to iterate, but update/delete is more
	sync.RWMutex
	visitors map[string]*Visitor
}

func NewVisitorManager() *VisitorManager {
	return &VisitorManager{visitors: make(map[string]*Visitor)}
}

func (v *VisitorManager) Add(id string) *Visitor {
	vst, ok := v.Get(id)
	if !ok {
		vst = NewVisitor()
		v.Lock()
		v.visitors[id] = vst
		v.Unlock()

	}
	return vst
}

func (v *VisitorManager) Get(id string) (*Visitor, bool) {
	v.RLock()
	vst, ok := v.visitors[id]
	v.RUnlock()
	return vst, ok
}

type AppSensor struct {
	policies *PolicyManager
	visitors *VisitorManager
}

type Option struct {
	Policies []Policy
}

func NewAppSensor(opt Option) *AppSensor {
	p := NewPolicyManager()
	p.Add(opt.Policies...)
	return &AppSensor{
		policies: p,
		visitors: NewVisitorManager(),
	}
}
func (aps *AppSensor) Allow(id string) bool {
	vst, ok := aps.visitors.Get(id)
	// Does not exist, which means it has not been penalized...
	if !ok {
		return true
	}
	return time.Now().After(vst.penalizedUntil)
}
func (aps *AppSensor) Penalize(id string, code EventCode) bool {
	pol, _ := aps.policies.Get(code)
	vst := aps.visitors.Add(id)
	vst.Add(code)
	if penalized := vst.HasCount(code, pol.Period, pol.Threshold); penalized {
		vst.Penalize(pol.PenalizedDuration)
		return true
	}
	return false
}
func main() {

	opt := Option{
		Policies: []Policy{
			Policy{
				Code:              BadPassword,
				Period:            time.Second,     // Within 1 seconds,
				Threshold:         3,               // If there is 3 event,
				PenalizedDuration: 5 * time.Second, // The user is locked for 5 seconds
			},
			Policy{
				Code:              BadRequest,
				Period:            time.Second,     // Within 1 seconds,
				Threshold:         1,               // If there is one event,
				PenalizedDuration: 1 * time.Second, // The user is locked for 1 second
			},
		},
	}
	aps := NewAppSensor(opt)

	allow := aps.Allow("0.0.0.0")

	log.Println("can the ip pass?", allow)
	log.Println(aps.Penalize("0.0.0.0", BadPassword))
	log.Println(aps.Penalize("0.0.0.0", BadPassword))
	log.Println(aps.Penalize("0.0.0.0", BadPassword))

	allow = aps.Allow("0.0.0.0")
	log.Println("can the ip pass?", allow)
	time.Sleep(3 * time.Second)

	allow = aps.Allow("0.0.0.0")
	log.Println("can the ip pass?", allow)
	time.Sleep(3 * time.Second)

	allow = aps.Allow("0.0.0.0")
	log.Println("can the ip pass?", allow)
	log.Println(aps.Penalize("0.0.0.0", BadPassword))
	log.Println(aps.Penalize("0.0.0.1", BadRequest))
	allow = aps.Allow("0.0.0.1")
	log.Println("can the ip pass?", allow)
	log.Println(aps.Penalize("0.0.0.1", BadRequest))
	log.Println(aps.Penalize("0.0.0.1", BadRequest))
	time.Sleep(4 * time.Second)
	allow = aps.Allow("0.0.0.1")
	log.Println("can the ip pass?", allow)
}
```
