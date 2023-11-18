CREATE TABLE IF NOT EXISTS categories (
    id bigserial PRIMARY KEY,
    "name" varchar(20) not null,
    description text not null,
    "alpha2" varchar(2) not null,
    "alpha3" varchar(3) not null
);
