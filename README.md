# Project manifest


## Concept

Let's test out the concept of rewarding children for performing chores.

We are focusing on gamification here.

- First up we are creating an admin UI for creating blueprints 


## Todos

- Add a main view
- Add a way to add Routine blueprints with one or more Routine Blueprint Chore


## Technical rules

- Use modern Go (generics, minimal deps, idiomatic)
- Minimal dependencies
- This is a hobby project. Do not worry about:
    - Backward compatibility
    - Breaking API contracts
- Server-side HTML rendering
- Modern CSS (Grid, Variables, etc. — open to discovery)
- Love for SQLite3
- Repetition is okay. DRY is dead. SOLID is dead.
- HTMX is cool. HTML is cool.

## Dependencies

- Tailwind is the root of all evil
- Use https://github.com/mattn/go-sqlite3
- Use https://templ.guide/ as template rendering



### Migrations

We are writing our own mini migration engine.

there should be a migrations dir. 

we are naming our migrations 1 and 2 and 3 and so on. 

If the migrations table has migration 1 in it, but 1.sql 2.sql, 3.sql exist in the migrations dir, then 2 must be applied in a transaction, then 3.sql

If 2 fails then do not apply 3.



## Data model

### User
- ID int64 Primary key
- Created time.Time
- Modified time.Time
- Name string required
- Password string required (not serialized to JSON)

### Chore
- ID int64 Primary key
- Created time.Time
- Modified time.Time
- Name string
- DefaultPoints int
- Image string

### Routine Blueprint
- ID int64 Primary key
- Created time.Time
- Modified time.Time
- Name string
- ToBeCompletedBy string
- AllowMultipleInstancesPerDay bool
- Recurrence enum (Daily, Weekly, Weekday)
- Image string

### Routine Blueprint Chore
- ID int64 Primary key
- Created time.Time
- Modified time.Time
- RoutineBlueprintID int64 (FK to routine blueprint)
- ChoreID int64 (FK to chore)
- Image string
- Chore *Chore (convenience field, not stored in DB)

### Routine
- ID int64 Primary key
- Created time.Time
- Modified time.Time
- OwnerID int64 (FK to User)
- RoutineBlueprintID sql.NullInt64 (optional FK to routine blueprint)
- ImageUrl string
- Owner *User (convenience field, not stored in DB)

### ChoreRoutine
- ID int64 Primary key
- Created time.Time
- Modified time.Time
- CompletedAt *time.Time (nullable)
- CompletedByID *int64 (nullable FK to user)
- PointsAwarded int
- RoutineID int64 (FK to Routine)
- ChoreID int64 (FK to Chore)
- CompletedBy *User (convenience field, not stored in DB)
- Routine *Routine (convenience field, not stored in DB)
- Chore *Chore (convenience field, not stored in DB)

## Migration
- ID int Primary key
- applied_at datetime
- migration_id Unique positive integer (sequential)




