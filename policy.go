package main

import "time"

type Policy struct {
	Code              EventCode
	Threshold         int
	Period            time.Duration // Period is the time range where the threshold is honoured.
	PenalizedDuration time.Duration
}
