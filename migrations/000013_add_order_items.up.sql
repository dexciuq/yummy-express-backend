CREATE TABLE if not exists order_items(
    id bigserial,
    order_id bigint,
    product_id bigint,
    quantity double precision,
    total bigint,
    CONSTRAINT order_id FOREIGN KEY (order_id)
    REFERENCES orders (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
    NOT VALID,
    CONSTRAINT product_id FOREIGN KEY (product_id)
    REFERENCES products (id) MATCH SIMPLE
                            ON UPDATE NO ACTION
                            ON DELETE NO ACTION
    NOT VALID
);

