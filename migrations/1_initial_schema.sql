-- Create migrations table
CREATE TABLE IF NOT EXISTS migrations (
    id INTEGER PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    migration_id INTEGER UNIQUE NOT NULL
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    password TEXT NOT NULL
);

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
    name TEXT NOT NULL,
    image TEXT NOT NULL,
    allow_multiple_instances_per_day BOOLEAN NOT NULL DEFAULT 0,
    recurrence TEXT CHECK (recurrence IN ('Daily', 'Weekly', "Weekday") OR recurrence IS NULL)
);

-- Create routine_blueprint_chores table
CREATE TABLE IF NOT EXISTS routine_blueprint_chores (
    id INTEGER PRIMARY KEY,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    routine_blueprint_id INTEGER NOT NULL,
    chore_id INTEGER NOT NULL,
    FOREIGN KEY (routine_blueprint_id) REFERENCES routine_blueprints(id),
    FOREIGN KEY (chore_id) REFERENCES chores(id)
);

-- Create routines table
CREATE TABLE IF NOT EXISTS routines (
    id INTEGER PRIMARY KEY,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    owner_id INTEGER NOT NULL,
    routine_blueprint_id INTEGER,
    FOREIGN KEY (owner_id) REFERENCES users(id),
    FOREIGN KEY (routine_blueprint_id) REFERENCES routine_blueprints(id)
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

-- Seed chores
INSERT INTO chores (name, default_points, image) VALUES
('Spis morgenmad', 10, 'img/breakfast.png'),
('Tag tøj på', 10, 'img/get-dressed.png'),
('Børst tænder (morgen)', 5, 'img/brush-teeth.png'),
('Pak madkasse', 15, 'img/lunch-box.png'),
('Kom ud af døren', 5, 'img/door.png'),
('Vask hænder', 5, 'img/wash-hands.png'),
('Dæk bordet', 15, 'img/set-table.png'),
('Hjælp med madlavning', 25, 'img/prepare-dinner.png'),
('Børst tænder (aften)', 5, 'img/brush-teeth.png'),
('Tag nattøj på', 5, 'img/pyjamas.png'),
('Læs en bog', 20, 'img/read-story-father.png');

-- Seed routine blueprints
INSERT INTO routine_blueprints (name, to_be_completed_by, image, allow_multiple_instances_per_day, recurrence) VALUES
('Morgen', '08:00:00', 'img/morning.png', 0, 'Weekday'),
('Eftermiddag', '17:00:00', 'img/afternoon.png', 0, 'Weekday'),
('Aften', '20:00:00', 'img/bedtime.png', 0, 'Weekday');

-- Seed routine_blueprint_chores
-- Morning (id=1)
INSERT INTO routine_blueprint_chores (routine_blueprint_id, chore_id) VALUES
(1, 1), -- Spis morgenmad
(1, 2), -- Tag tøj på
(1, 3), -- Børst tænder (morgen)
(1, 4), -- Pak madkasse
(1, 5); -- Kom ud af døren

-- Afternoon (id=2)
INSERT INTO routine_blueprint_chores (routine_blueprint_id, chore_id) VALUES
(2, 6), -- Vask hænder
(2, 7), -- Dæk bordet
(2, 8); -- Hjælp med madlavning

-- Bedtime (id=3)
INSERT INTO routine_blueprint_chores (routine_blueprint_id, chore_id) VALUES
(3, 9), -- Børst tænder (aften)
(3, 10), -- Tag nattøj på
(3, 11); -- Læs en bog

-- Seed users
INSERT INTO users (id, name, password) VALUES (1, 'werner', '');

-- Seed routines
-- INSERT INTO routines (owner_id, routine_blueprint_id) VALUES
-- (1, 1), -- Morning
-- (1, 2), -- Afternoon
-- (1, 3); -- Bedtime

