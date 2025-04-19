package models

import "time"

type Routine struct {
	ID       int64     `json:"id"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
	OwnerID  int64     `json:"owner_id"`
	ImageUrl string    `json:"image_url,omitempty"`

	// These fields are not stored in the database but can be populated for convenience
	Owner *User `json:"owner,omitempty"`
}
