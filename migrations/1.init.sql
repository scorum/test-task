CREATE TABLE account (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    account_id TEXT,
    brand_id TEXT,
    currency TEXT,
    balance double precision,

    PRIMARY KEY (id),
    UNIQUE (account_id, brand_id)
);