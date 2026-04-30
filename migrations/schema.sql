CREATE TABLE wallets (
    id BIGSERIAL PRIMARY KEY,
    balance BIGINT NOT NULL CHECK (balance >= 0)
);

CREATE TABLE transfers (
    id BIGSERIAL PRIMARY KEY,
    from_wallet_id BIGINT NOT NULL REFERENCES wallets(id),
    to_wallet_id BIGINT NOT NULL REFERENCES wallets(id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    state TEXT NOT NULL CHECK (state IN ('PENDING', 'PROCESSED', 'FAILED')),
    idempotency_key TEXT UNIQUE
);
CREATE TABLE ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL REFERENCES wallets(id),
    transfer_id BIGINT NOT NULL REFERENCES transfers(id),
    type TEXT NOT NULL CHECK (type IN ('debit', 'credit')),
    amount BIGINT NOT NULL CHECK (amount > 0)
);

CREATE TABLE idempotency_records (
    idempotency_key TEXT PRIMARY KEY,
    request_hash TEXT NOT NULL,

    status TEXT DEFAULT 'IN_PROGRESS', -- IN_PROGRESS | COMPLETED | FAILED

    response_payload TEXT DEFAULT '',
    status_code INT,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);