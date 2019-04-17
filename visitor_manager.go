package main

import (
	"context"
	"log"
	"sync"
	"time"
)

type VisitorManager struct {
	// Slice would be faster to iterate, but update/delete is more
	sync.RWMutex
	visitors map[string]*Visitor
}

func NewVisitorManager() *VisitorManager {
	return &VisitorManager{visitors: make(map[string]*Visitor)}
}
func (v *VisitorManager) Start(policies *PolicyManager) func(context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	t := time.NewTicker(policies.Hd().Period)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				log.Println("stopped visitor manager")
				return
			case <-t.C:
				for _, vst := range v.visitors {
					vst.ClearBefore(policies)
				}
			}
		}

	}()
	return func(ctx context.Context) {
		sig := make(chan interface{})
		go func() {
			close(done)
			wg.Wait()
			close(sig)
		}()
		select {
		case <-sig:
			return
		case <-ctx.Done():
			return
		}
	}
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
