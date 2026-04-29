CREATE TABLE wallets (
    id BIGSERIAL PRIMARY KEY,
    balance BIGINT NOT NULL CHECK (balance >= 0)
);

CREATE TABLE transfers (
    id BIGSERIAL PRIMARY KEY,
    from_wallet_id BIGINT,
    to_wallet_id BIGINT,
    amount BIGINT,
    state TEXT,
    idempotency_key TEXT UNIQUE
);

CREATE TABLE ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT,
    transfer_id BIGINT,
    type TEXT,
    amount BIGINT
);

CREATE TABLE idempotency_records (
    idempotency_key TEXT PRIMARY KEY,
    request_hash TEXT,
    response_payload TEXT,
    status_code INT
);