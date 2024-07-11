create table
  IF not exists trips (
    "id" uuid primary KEY not null default gen_random_uuid (),
    "destination" varchar(255) not null,
    "owner_email" varchar(255) not null,
    "owner_name" varchar(255) not null,
    "is_confirmed" boolean not null default false,
    "starts_at" timestamp not null,
    "ends_at" timestamp not null
  );

---- create above / drop below ----
drop table IF exists trips;