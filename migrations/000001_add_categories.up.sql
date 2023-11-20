CREATE TABLE IF NOT EXISTS categories (
    id bigserial PRIMARY KEY,
    "name" varchar(20) not null UNIQUE,
    description text not null,
    image text not null
);