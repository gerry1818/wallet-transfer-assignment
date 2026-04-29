-- Clear existing data (optional, useful for repeatable runs)
TRUNCATE TABLE ledger_entries RESTART IDENTITY CASCADE;
TRUNCATE TABLE transfers RESTART IDENTITY CASCADE;
TRUNCATE TABLE wallets RESTART IDENTITY CASCADE;
TRUNCATE TABLE idempotency_records RESTART IDENTITY CASCADE;

-- Insert sample wallets
INSERT INTO wallets (balance) VALUES (1000); -- id = 1
INSERT INTO wallets (balance) VALUES (500);  -- id = 2
