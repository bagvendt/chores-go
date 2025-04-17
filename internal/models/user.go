package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
	Name      string    `json:"name"`
	Password  string    `json:"-"` // Password is never serialized to JSON
} 