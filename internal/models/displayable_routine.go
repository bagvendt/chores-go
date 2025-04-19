package models

import "time"

// SourceType identifies whether a DisplayableRoutine comes from a database record or a blueprint
type SourceType string

const (
	DatabaseSource  SourceType = "database"
	BlueprintSource SourceType = "blueprint"
)

// DisplayableRoutine represents a routine that can be displayed to a user
// It can be either a concrete routine from the database or a virtual routine
// generated from a RoutineBlueprint
type DisplayableRoutine struct {
	// Fields common to both sources
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	ToBeCompletedBy string `json:"to_be_completed_by"`
	ImageUrl        string `json:"image_url,omitempty"`
	OwnerID         int64  `json:"owner_id"`
	Owner           *User  `json:"owner,omitempty"`

	// Metadata
	SourceType  SourceType `json:"source_type"`
	BlueprintID *int64     `json:"blueprint_id,omitempty"` // Only present for blueprint-sourced routines

	// If from a database source, includes these fields
	Created  *time.Time `json:"created,omitempty"`
	Modified *time.Time `json:"modified,omitempty"`

	// Information about chores in this routine
	ChoreCount      int `json:"chore_count"`
	CompletedChores int `json:"completed_chores"`

	// Original source objects (not serialized to JSON)
	FromRoutine   *Routine          `json:"-"`
	FromBlueprint *RoutineBlueprint `json:"-"`
}

// IsComplete returns true if all chores in the routine are complete
func (r *DisplayableRoutine) IsComplete() bool {
	return r.ChoreCount > 0 && r.CompletedChores == r.ChoreCount
}

// CompletionPercentage returns the percentage of completed chores (0-100)
func (r *DisplayableRoutine) CompletionPercentage() int {
	if r.ChoreCount == 0 {
		return 0
	}
	return (r.CompletedChores * 100) / r.ChoreCount
}
