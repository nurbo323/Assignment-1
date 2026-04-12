CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL UNIQUE,
    transaction_id TEXT NOT NULL UNIQUE,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status TEXT NOT NULL CHECK (status IN ('Authorized', 'Declined'))
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
