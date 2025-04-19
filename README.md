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
- ID Int Primary key
- Created datetime
- Modified datetime
- Name str required
- Password str required - Hash salt iterations something clever


### Chore
- ID Int Primary key
- Created datetime
- Modified datetime
- Name
- Default points positive int
- Image - str


### Routine Blueprint
- ID Int Primary key
- Created datetime
- Modified datetime
- to_be_completed_by: time_of_day
- allow_multiple_instances_per_day
- Recurrence: Nullable enum. Daily, Weekly, Weekday


### Routine Blueprint Chore
- ID Int Primary key
- Created datetime
- Modified datetime
- FK to routine blueprint
- FK to chore
- Image : str

 

### Routine
- ID Int Primary key
- Created datetime
- Modified datetime
- Name
- to_be_completed_by: time_of_day
- owner: FK to User


### ChoreRoutine
- ID Int Primary key
- Created datetime
- Modified datetime
- Completed_at nullable datetime
- Completed_by fk to user
- Points_awarded: positive int

## Migration
- ID Int Primary key
- applied_at datetime
- migration_id: Unique positive integer (sequential)




