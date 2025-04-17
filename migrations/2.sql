-- Create chores table
CREATE TABLE IF NOT EXISTS chores (
    id INTEGER PRIMARY KEY,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    default_points INTEGER NOT NULL CHECK (default_points > 0),
    image TEXT
);

-- Create routine_blueprints table
CREATE TABLE IF NOT EXISTS routine_blueprints (
    id INTEGER PRIMARY KEY,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    to_be_completed_by TEXT NOT NULL,
    allow_multiple_instances_per_day BOOLEAN NOT NULL DEFAULT 0,
    recurrence TEXT CHECK (recurrence IN ('Daily', 'Weekly') OR recurrence IS NULL)
);

-- Create routine_blueprint_chores table
CREATE TABLE IF NOT EXISTS routine_blueprint_chores (
    id INTEGER PRIMARY KEY,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    routine_blueprint_id INTEGER NOT NULL,
    chore_id INTEGER NOT NULL,
    image TEXT,
    FOREIGN KEY (routine_blueprint_id) REFERENCES routine_blueprints(id),
    FOREIGN KEY (chore_id) REFERENCES chores(id)
);

-- Create routines table
CREATE TABLE IF NOT EXISTS routines (
    id INTEGER PRIMARY KEY,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    to_be_completed_by TEXT NOT NULL,
    owner_id INTEGER NOT NULL,
    FOREIGN KEY (owner_id) REFERENCES users(id)
);

-- Create chore_routines table
CREATE TABLE IF NOT EXISTS chore_routines (
    id INTEGER PRIMARY KEY,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    completed_by INTEGER,
    points_awarded INTEGER CHECK (points_awarded > 0),
    routine_id INTEGER NOT NULL,
    chore_id INTEGER NOT NULL,
    FOREIGN KEY (completed_by) REFERENCES users(id),
    FOREIGN KEY (routine_id) REFERENCES routines(id),
    FOREIGN KEY (chore_id) REFERENCES chores(id)
);

-- Insert migration record
INSERT INTO migrations (migration_id) VALUES (2); 