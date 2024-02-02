CREATE TABLE if not exists order_items(
    id bigserial,
    user_id bigint,
    product_id bigint,
    quantity double precision,
    total bigint,
    CONSTRAINT user_id FOREIGN KEY (user_id)
    REFERENCES users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
    NOT VALID,
    CONSTRAINT product_id FOREIGN KEY (product_id)
    REFERENCES products (id) MATCH SIMPLE
                            ON UPDATE NO ACTION
                            ON DELETE NO ACTION
    NOT VALID
);

