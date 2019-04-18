package main

import "time"

type Policy struct {
	// TODO: Change this to detection point.
	Code              EventCode
	Threshold         int           // Threshold can also be in percentage.
	Period            time.Duration // Period is the time range where the threshold is honoured.
	PenalizedDuration time.Duration
	// Response
}
