create table
  IF not exists links (
    "id" uuid primary KEY not null default gen_random_uuid(),
    "trip_id" uuid not null,
    "title" varchar(255) not null,
    "url" varchar(255) not null default false,
    foreign KEY (trip_id) references trips (id) on update CASCADE on delete CASCADE
  );

---- create above / drop below ----
drop table IF exists links;