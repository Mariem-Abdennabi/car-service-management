CREATE TABLE customers (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    phone      TEXT        NOT NULL,
    city       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
