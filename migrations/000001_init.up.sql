-- up migration for creating the accounts schema, accounts table, transactions table, and timestamp trigger

-- 1. Create the schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS accounts;

-- 2. Create a reusable function in the 'accounts' schema to update the timestamp
CREATE OR REPLACE FUNCTION accounts.trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 3. Create the table with created_at and updated_at columns
CREATE TABLE IF NOT EXISTS accounts.accounts (
    id BIGINT PRIMARY KEY,
    balance NUMERIC(19, 4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Create the trigger to call the function before any update on the table
DROP TRIGGER IF EXISTS set_timestamp ON accounts.accounts;
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON accounts.accounts
FOR EACH ROW
EXECUTE PROCEDURE accounts.trigger_set_timestamp();

-- 5. Create the transactions table
CREATE TABLE IF NOT EXISTS accounts.transactions (
    id BIGSERIAL PRIMARY KEY,
    source_account_id BIGINT NOT NULL,
    destination_account_id BIGINT NOT NULL,
    amount NUMERIC(19, 4) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);