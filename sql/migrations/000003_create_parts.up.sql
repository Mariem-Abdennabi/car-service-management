CREATE TABLE parts (
    id               BIGSERIAL PRIMARY KEY,
    reference        TEXT        NOT NULL,
    name             TEXT        NOT NULL,
    -- Money as a whole number of millimes (1 dinar = 1000), never a float:
    -- 0.1 + 0.2 is not 0.3 in binary floating point, and prices get added up.
    price_millimes   INTEGER     NOT NULL CHECK (price_millimes >= 0),
    -- Enforced here as well as in the form: once repair jobs consume parts, the
    -- database is the only place that can refuse to go below zero under
    -- concurrent updates.
    quantity_on_hand INTEGER     NOT NULL DEFAULT 0 CHECK (quantity_on_hand >= 0),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
