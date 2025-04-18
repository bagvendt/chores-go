package database

import (
	"database/sql"
	"time"

	"github.com/bagvendt/chores/internal/models"
)

// GetBlueprints returns all routine blueprints
func GetBlueprints() ([]models.RoutineBlueprint, error) {
	rows, err := DB.Query(`
		SELECT id, created, modified, name, to_be_completed_by, allow_multiple_instances_per_day, recurrence
		FROM routine_blueprints
		ORDER BY created DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blueprints []models.RoutineBlueprint
	for rows.Next() {
		var blueprint models.RoutineBlueprint
		var createdStr, modifiedStr string
		var recurrence string

		if err := rows.Scan(
			&blueprint.ID,
			&createdStr,
			&modifiedStr,
			&blueprint.Name,
			&blueprint.ToBeCompletedBy,
			&blueprint.AllowMultipleInstancesPerDay,
			&recurrence,
		); err != nil {
			return nil, err
		}

		blueprint.Created, _ = time.Parse(time.RFC3339, createdStr)
		blueprint.Modified, _ = time.Parse(time.RFC3339, modifiedStr)
		blueprint.Recurrence = models.RecurrenceType(recurrence)

		blueprints = append(blueprints, blueprint)
	}

	return blueprints, nil
}

// GetBlueprint returns a single routine blueprint by ID along with its chores
func GetBlueprint(id int64) (*models.RoutineBlueprint, []models.RoutineBlueprintChore, error) {
	// Get the blueprint
	var blueprint models.RoutineBlueprint
	var createdStr, modifiedStr string
	var recurrence string

	err := DB.QueryRow(`
		SELECT id, created, modified, name, to_be_completed_by, allow_multiple_instances_per_day, recurrence
		FROM routine_blueprints
		WHERE id = ?
	`, id).Scan(
		&blueprint.ID,
		&createdStr,
		&modifiedStr,
		&blueprint.Name,
		&blueprint.ToBeCompletedBy,
		&blueprint.AllowMultipleInstancesPerDay,
		&recurrence,
	)
	if err != nil {
		return nil, nil, err
	}

	blueprint.Created, _ = time.Parse(time.RFC3339, createdStr)
	blueprint.Modified, _ = time.Parse(time.RFC3339, modifiedStr)
	blueprint.Recurrence = models.RecurrenceType(recurrence)

	// Get the chores for this blueprint
	rows, err := DB.Query(`
		SELECT 
			rbc.id, rbc.created, rbc.modified, rbc.routine_blueprint_id, rbc.chore_id, rbc.image,
			c.id, c.name, c.default_points, c.image
		FROM routine_blueprint_chores rbc
		JOIN chores c ON rbc.chore_id = c.id
		WHERE rbc.routine_blueprint_id = ?
	`, id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var chores []models.RoutineBlueprintChore
	for rows.Next() {
		var chore models.RoutineBlueprintChore
		var choreCreatedStr, choreModifiedStr string
		var choreImage sql.NullString
		var choreObj models.Chore
		var choreObjImage sql.NullString

		if err := rows.Scan(
			&chore.ID,
			&choreCreatedStr,
			&choreModifiedStr,
			&chore.RoutineBlueprintID,
			&chore.ChoreID,
			&choreImage,
			&choreObj.ID,
			&choreObj.Name,
			&choreObj.DefaultPoints,
			&choreObjImage,
		); err != nil {
			return nil, nil, err
		}

		chore.Created, _ = time.Parse(time.RFC3339, choreCreatedStr)
		chore.Modified, _ = time.Parse(time.RFC3339, choreModifiedStr)
		if choreImage.Valid {
			chore.Image = choreImage.String
		} else {
			chore.Image = ""
		}
		if choreObjImage.Valid {
			choreObj.Image = choreObjImage.String
		} else {
			choreObj.Image = ""
		}
		chore.Chore = &choreObj

		chores = append(chores, chore)
	}

	return &blueprint, chores, nil
}

// CreateBlueprint creates a new routine blueprint
func CreateBlueprint(blueprint *models.RoutineBlueprint, choreIDs []int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO routine_blueprints (
			name,
			to_be_completed_by, 
			allow_multiple_instances_per_day,
			recurrence
		) VALUES (?, ?, ?, ?)
	`,
		blueprint.Name,
		blueprint.ToBeCompletedBy,
		blueprint.AllowMultipleInstancesPerDay,
		string(blueprint.Recurrence),
	)
	if err != nil {
		return err
	}

	blueprintID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	blueprint.ID = blueprintID

	// Add chores to the blueprint
	for _, choreID := range choreIDs {
		_, err = tx.Exec(`
			INSERT INTO routine_blueprint_chores (
				routine_blueprint_id,
				chore_id
			) VALUES (?, ?)
		`,
			blueprintID,
			choreID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateBlueprint updates an existing routine blueprint
func UpdateBlueprint(blueprint *models.RoutineBlueprint, choreIDs []int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update blueprint
	_, err = tx.Exec(`
		UPDATE routine_blueprints
		SET name = ?,
			to_be_completed_by = ?,
			allow_multiple_instances_per_day = ?,
			recurrence = ?,
			modified = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		blueprint.Name,
		blueprint.ToBeCompletedBy,
		blueprint.AllowMultipleInstancesPerDay,
		string(blueprint.Recurrence),
		blueprint.ID,
	)
	if err != nil {
		return err
	}

	// Remove existing chores
	_, err = tx.Exec(`DELETE FROM routine_blueprint_chores WHERE routine_blueprint_id = ?`, blueprint.ID)
	if err != nil {
		return err
	}

	// Add new chores
	for _, choreID := range choreIDs {
		_, err = tx.Exec(`
			INSERT INTO routine_blueprint_chores (
				routine_blueprint_id,
				chore_id
			) VALUES (?, ?)
		`,
			blueprint.ID,
			choreID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeleteBlueprint deletes a routine blueprint and its associated chores
func DeleteBlueprint(id int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete associated chores first
	_, err = tx.Exec(`DELETE FROM routine_blueprint_chores WHERE routine_blueprint_id = ?`, id)
	if err != nil {
		return err
	}

	// Delete the blueprint
	_, err = tx.Exec(`DELETE FROM routine_blueprints WHERE id = ?`, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
