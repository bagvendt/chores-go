package models

import "time"

type Chore struct {
	ID            int64     `json:"id"`
	Created       time.Time `json:"created"`
	Modified      time.Time `json:"modified"`
	Name          string    `json:"name"`
	DefaultPoints int       `json:"default_points"`
	Image         string    `json:"image,omitempty"`
} 