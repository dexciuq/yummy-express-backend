CREATE TABLE IF NOT EXISTS roles (
    id bigserial PRIMARY KEY,
    "name" varchar(20) not null UNIQUE,
    description text not null
);