create table
  IF not exists participants (
    "id" uuid primary KEY not null default gen_random_uuid(),
    "trip_id" uuid not null,
    "email" varchar(255) not null,
    "is_confirmed" boolean not null default false,
    foreign KEY (trip_id) references trips (id) on update CASCADE on delete CASCADE
  );

---- create above / drop below ----
drop table IF exists participants;