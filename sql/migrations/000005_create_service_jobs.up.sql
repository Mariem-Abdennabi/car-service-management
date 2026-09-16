CREATE TABLE service_jobs (
    id              BIGSERIAL PRIMARY KEY,
    -- RESTRICT, like vehicles.customer_id: a vehicle with service history cannot
    -- be deleted out from under it.
    vehicle_id      BIGINT      NOT NULL REFERENCES vehicles (id) ON DELETE RESTRICT,
    -- TEXT with a CHECK rather than a PostgreSQL ENUM. Both refuse a value that is
    -- not in the list; changing a CHECK is ordinary SQL in a migration, while
    -- altering an enum type has its own rules about what may run in a transaction.
    status          TEXT        NOT NULL DEFAULT 'received'
        CHECK (status IN ('received', 'in_progress', 'completed', 'cancelled')),
    -- What the customer reported. Free text: the mechanic's own words are the
    -- record, and no list of categories survives contact with a real workshop.
    description     TEXT        NOT NULL,
    -- Millimes, never a float — the same rule as parts.price_millimes.
    labour_millimes INTEGER     NOT NULL DEFAULT 0 CHECK (labour_millimes >= 0),
    opened_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Every job lookup on a customer's page goes through the vehicle, and PostgreSQL
-- does not index a foreign key automatically.
CREATE INDEX service_jobs_vehicle_id_idx ON service_jobs (vehicle_id);
