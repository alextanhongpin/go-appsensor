package main

import (
	"context"
	"log"
	"time"
)

type EventCode uint

const (
	BadPassword EventCode = iota
	BadRequest
)

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
func (aps *AppSensor) Start() func(context.Context) {
	return aps.visitors.Start(aps.policies)
}
func (aps *AppSensor) Allow(id string) bool {
	vst, ok := aps.visitors.Get(id)
	// Does not exist, which means it has not been penalized...
	if !ok {
		return true
	}
	return time.Now().After(vst.penalizedUntil)
}

// NOTE: This API is oversimplified, there are probably more information that needs to be gathered here such as:
// - detection points
// - client_id (client referring to the application)
// - event date/time
// - url path
// - http method
// - source ip address
// - source user agent
// - query string
// - bytes transferred
// - response status code
// LOCATION:
// - host
// - service/application name
// - entry point
// APP SENSOR DETECTION
// - sensor id
// - sensor location
// - appsensor detection point
// - description
// - message
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
				Period:            2 * time.Second, // Within 2 seconds,
				Threshold:         3,               // If there is 3 event,
				PenalizedDuration: 5 * time.Second, // The user is locked for 5 seconds
			},
			Policy{
				Code:              BadRequest,
				Period:            time.Second,     // Within 1 second,
				Threshold:         1,               // If there is one event,
				PenalizedDuration: 1 * time.Second, // The user is locked for 1 second
			},
		},
	}
	aps := NewAppSensor(opt)
	shutdown := aps.Start()

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
	log.Println("can the ip 0.0.0.1 pass?", allow)
	log.Println(aps.Penalize("0.0.0.1", BadRequest))
	log.Println(aps.Penalize("0.0.0.1", BadRequest))
	time.Sleep(2 * time.Second)
	allow = aps.Allow("0.0.0.1")
	log.Println("can the ip 0.0.0.1 pass?", allow)
	time.Sleep(2 * time.Second)
	allow = aps.Allow("0.0.0.1")
	log.Println("can the ip 0.0.0.1 pass?", allow)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	shutdown(ctx)
	log.Println("terminated gracefully")
}
