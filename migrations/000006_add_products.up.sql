CREATE TABLE if not exists products (
    id bigserial NOT NULL PRIMARY KEY,
    "name" character varying(20) NOT NULL,
    price bigint,
    description text,
    upc character varying,
    quantity bigint,
    image text,
    created_at timestamp(0) with time zone not null default NOW(),
    category_id bigint,
    discount_id bigint,
    unit_id bigint,
    brand_id bigint,
    country_id bigint,
    step double precison,
    CONSTRAINT category_id FOREIGN KEY (category_id)
        REFERENCES categories (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID,
    CONSTRAINT discount FOREIGN KEY (discount_id)
        REFERENCES discounts (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID,
    CONSTRAINT unit FOREIGN KEY (unit_id)
        REFERENCES units (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID,
    CONSTRAINT brand FOREIGN KEY (brand_id)
        REFERENCES brands (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID,
    CONSTRAINT country_id FOREIGN KEY (country_id)
        REFERENCES countries (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID
);
