CREATE TABLE vehicles (
    id          BIGSERIAL PRIMARY KEY,
    customer_id BIGINT      NOT NULL REFERENCES customers (id) ON DELETE RESTRICT,
    plate       TEXT        NOT NULL,
    make        TEXT        NOT NULL,
    model       TEXT        NOT NULL,
    year        INTEGER     NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Every vehicle lookup goes through its customer, and PostgreSQL does not index a
-- foreign key automatically.
CREATE INDEX vehicles_customer_id_idx ON vehicles (customer_id);
