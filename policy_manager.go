package main

import (
	"sort"
	"sync"
	"time"
)

// PolicyManager holds the policies.
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

func (p *PolicyManager) Hd() Policy {
	var pol Policy
	p.RLock()
	if len(p.policies) == 0 {
		pol = p.plain
	} else {
		pol = p.policies[0]
	}
	p.RUnlock()
	return pol
}

func (p *PolicyManager) Add(pol ...Policy) {
	p.Lock()
	p.policies = append(p.policies, pol...)
	// Sort it by smallest to largest period.
	sort.Slice(p.policies, func(i, j int) bool {
		return p.policies[i].Period < p.policies[i].Period
	})
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
