package models

import "time"

type Routine struct {
	ID              int64     `json:"id"`
	Created         time.Time `json:"created"`
	Modified        time.Time `json:"modified"`
	Name            string    `json:"name"`
	ToBeCompletedBy string    `json:"to_be_completed_by"`
	OwnerID         int64     `json:"owner_id"`
	
	// These fields are not stored in the database but can be populated for convenience
	Owner           *User     `json:"owner,omitempty"`
} 