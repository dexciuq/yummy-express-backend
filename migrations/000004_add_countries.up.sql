CREATE TABLE IF NOT EXISTS countries (
    id bigserial PRIMARY KEY,
    "name" varchar(20) not null UNIQUE,
    description text not null,
    "alpha2" varchar(2) not null UNIQUE,
    "alpha3" varchar(3) not null UNIQUE
);
