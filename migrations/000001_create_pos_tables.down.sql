-- Down migration for POS module
-- Drop all POS module tables in reverse order

DROP TABLE IF EXISTS pos_employees CASCADE;
DROP TABLE IF EXISTS pos_shifts CASCADE;
DROP TABLE IF EXISTS pos_taxes CASCADE;
DROP TABLE IF EXISTS pos_discounts CASCADE;
DROP TABLE IF EXISTS pos_customers CASCADE;
DROP TABLE IF EXISTS pos_products CASCADE;
DROP TABLE IF EXISTS pos_cash_drawer_transactions CASCADE;
DROP TABLE IF EXISTS pos_cash_drawers CASCADE;
DROP TABLE IF EXISTS pos_registers CASCADE;
DROP TABLE IF EXISTS pos_receipts CASCADE;
DROP TABLE IF EXISTS pos_payments CASCADE;
DROP TABLE IF EXISTS pos_transaction_items CASCADE;
DROP TABLE IF EXISTS pos_transactions CASCADE;
DROP TABLE IF EXISTS pos_sessions CASCADE;
DROP TABLE IF EXISTS register_shifts CASCADE;
DROP TABLE IF EXISTS pos_registers CASCADE;

