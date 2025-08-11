-- down migration for the accounts schema and table

-- Drop the trigger from the table
DROP TRIGGER IF EXISTS set_timestamp ON accounts.accounts;

-- Drop the accounts table
DROP TABLE IF EXISTS accounts.accounts;

-- Drop the transactions table
DROP TABLE IF EXISTS accounts.transactions;

-- Drop the timestamp function
DROP FUNCTION IF EXISTS accounts.trigger_set_timestamp();

-- Drop the schema
DROP SCHEMA IF EXISTS accounts;