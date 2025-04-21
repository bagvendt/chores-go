-- Remove 'img/' prefix from image column in chores table
UPDATE chores
SET image = REPLACE(image, 'img/', '')
WHERE image LIKE 'img/%';

-- Remove 'img/' prefix from image column in routine_blueprints table
UPDATE routine_blueprints
SET image = REPLACE(image, 'img/', '')
WHERE image LIKE 'img/%';