package models

import "time"

type ChoreRoutine struct {
	ID            int64      `json:"id"`
	Created       time.Time  `json:"created"`
	Modified      time.Time  `json:"modified"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CompletedByID *int64     `json:"completed_by,omitempty"`
	PointsAwarded int        `json:"points_awarded"`
	RoutineID     int64      `json:"routine_id"`
	ChoreID       int64      `json:"chore_id"`
	
	// These fields are not stored in the database but can be populated for convenience
	CompletedBy   *User      `json:"completed_by_user,omitempty"`
	Routine       *Routine   `json:"routine,omitempty"`
	Chore         *Chore     `json:"chore,omitempty"`
}