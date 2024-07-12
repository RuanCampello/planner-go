-- name: InsertTrip :one
insert into
  trips (
    "destination",
    "owner_email",
    "owner_name",
    "starts_at",
    "ends_at"
  ) values ($1, $2, $3, $4, $5)
returning "id";

-- name: GetTrip :one
select
    "id", 
    "destination", 
    "owner_email", 
    "owner_name", 
    "is_confirmed", 
    "starts_at", 
    "ends_at"
from trips
where
    id = $1;

-- name: UpdateTrip :exec
UPDATE trips
SET 
    "destination" = $1,
    "ends_at" = $2,
    "starts_at" = $3,
    "is_confirmed" = $4
where
    id = $5;

-- name: GetParticipant :one
select
    "id", 
    "trip_id", 
    "email", 
    "is_confirmed"
from participants
where
    id = $1;

-- name: ConfirmParticipant :exec
select
    "id", 
    "trip_id", 
    "email", 
    "is_confirmed"
from participants
where
    id = $1;


-- name: GetParticipants :many
select
    "id", 
    "trip_id", 
    "email", 
    "is_confirmed"
from participants
where
    trip_id = $1;

-- name: InviteParticipantToTrip :one
INSERT INTO participants
    ( "trip_id", "email" ) VALUES
    ( $1, $2 )
RETURNING "id";

-- name: InviteParticipantsToTrip :copyfrom
insert into participants
    ( "trip_id", "email" ) values
    ( $1, $2 );

-- name: CreateActivity :one
insert into activities
    ( "trip_id", "title", "occurs_at" ) values
    ( $1, $2, $3 )
returning "id";

-- name: GetTripActivities :many
select
    "id", 
    "trip_id", 
    "title", 
    "occurs_at"
from activities
where
    trip_id = $1;

-- name: CreateTripLink :one
insert into links
    ( "trip_id", "title", "url" ) values
    ( $1, $2, $3 )
returning "id";

-- name: GetTripLinks :many
select
    "id", 
    "trip_id", 
    "title", 
    "url"
from links
where
    trip_id = $1;

