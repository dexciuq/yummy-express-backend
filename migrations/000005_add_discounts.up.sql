CREATE TABLE IF NOT EXISTS discounts (
    id bigserial PRIMARY KEY,
    "name" varchar(20) not null,
    description text not null,
    discount_percent int not null,
    created_at timestamp(0) with time zone not null default NOW(),
    started_at timestamp(0) with time zone not null,
    ended_at timestamp(0) with time zone not null
);