CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    item_name TEXT NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status TEXT NOT NULL CHECK (status IN ('Pending', 'Paid', 'Failed', 'Cancelled')),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS order_idempotency (
    idempotency_key TEXT PRIMARY KEY,
    request_hash TEXT NOT NULL,
    order_id TEXT NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
