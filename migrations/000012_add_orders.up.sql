CREATE TABLE if not exists orders(
    id bigserial NOT NULL PRIMARY KEY,
    user_id bigint,
    total bigint,
    address character varying(255),
    status_id bigint,
    created_at timestamp(0) with time zone not null default NOW(),
    delivered_at timestamp(0) with time zone,
    CONSTRAINT user_id FOREIGN KEY (user_id)
    REFERENCES users (id) MATCH SIMPLE
                         ON UPDATE NO ACTION
                         ON DELETE NO ACTION
    NOT VALID,
    CONSTRAINT status_id FOREIGN KEY (status_id)
    REFERENCES statuses (id) MATCH SIMPLE
                            ON UPDATE NO ACTION
                            ON DELETE NO ACTION
    NOT VALID
);

