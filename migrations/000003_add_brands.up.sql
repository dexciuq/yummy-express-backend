CREATE TABLE IF NOT EXISTS brands (
    id bigserial PRIMARY KEY,
    "name" varchar(20) not null UNIQUE,
    description text not null
);