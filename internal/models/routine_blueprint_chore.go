package models

import "time"

type RoutineBlueprintChore struct {
	ID                 int64     `json:"id"`
	Created            time.Time `json:"created"`
	Modified           time.Time `json:"modified"`
	RoutineBlueprintID int64     `json:"routine_blueprint_id"`
	ChoreID            int64     `json:"chore_id"`
	Image              string    `json:"image,omitempty"`

	// These fields are not stored in the database but can be populated for convenience
	Chore *Chore `json:"chore,omitempty"`
}
