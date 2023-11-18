CREATE TABLE IF NOT EXISTS units (
    id bigserial PRIMARY KEY,
    "name" varchar(20) not null,
    description text not null
);