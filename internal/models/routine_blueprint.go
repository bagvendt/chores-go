package models

import "time"

// RecurrenceType defines the possible recurrence types for a routine blueprint
type RecurrenceType string

const (
	Daily   RecurrenceType = "Daily"
	Weekly  RecurrenceType = "Weekly"
	Weekday RecurrenceType = "Weekday"
)

type RoutineBlueprint struct {
	ID                           int64          `json:"id"`
	Created                      time.Time      `json:"created"`
	Modified                     time.Time      `json:"modified"`
	Name                         string         `json:"name"`
	ToBeCompletedBy              string         `json:"to_be_completed_by"`
	AllowMultipleInstancesPerDay bool           `json:"allow_multiple_instances_per_day"`
	Recurrence                   RecurrenceType `json:"recurrence,omitempty"`
	Image                        string         `json:"image,omitempty"`
}
