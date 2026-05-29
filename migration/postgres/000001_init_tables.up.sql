BEGIN;

CREATE TABLE IF NOT EXISTS orders(
    id BIGINT NOT NULL UNIQUE,
    guid UUID NOT NULL PRIMARY KEY,
    user_guid UUID DEFAULT NULL,
    total_price DECIMAL(12,3) DEFAULT 0,
    delivery_price DECIMAL(12,3) DEFAULT 0,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items(
    id BIGINT NOT NULL UNIQUE,
    guid UUID NOT NULL PRIMARY KEY,
    order_guid UUID NOT NULL REFERENCES orders(guid) ON DELETE RESTRICT,
    product_guid UUID NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(12,3)
);
COMMIT;
