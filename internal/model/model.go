package model

import "time"

type Event struct {
	EventTimestamp time.Time `json:"event_timestamp"`
	Body           string    `json:"body"`
}

type Response struct {
	Status string `json:"status"`
}
