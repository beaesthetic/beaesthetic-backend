DROP INDEX IF EXISTS idx_wallet_operations_occurred_at;
DROP INDEX IF EXISTS idx_wallet_operations_credit_lot_id;
DROP INDEX IF EXISTS idx_wallet_operations_wallet_id;
DROP TABLE IF EXISTS wallet_operations;

DROP INDEX IF EXISTS idx_wallet_credit_lots_expires_at;
DROP INDEX IF EXISTS idx_wallet_credit_lots_wallet_id;
DROP TABLE IF EXISTS wallet_credit_lots;
