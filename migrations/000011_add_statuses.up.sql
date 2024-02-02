CREATE TABLE IF NOT EXISTS statuses (
    id bigserial PRIMARY KEY,
    "name" varchar(50) not null UNIQUE,
    description text not null
);
