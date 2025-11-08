package models

import "time"

type Job struct {
	ID        int
	Type      string
	Payload   string // or map[string]interface{} for JSON
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
