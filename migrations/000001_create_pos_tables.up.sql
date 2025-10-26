-- POS Module Database Schema
-- This migration creates all tables for the point of sale system
-- Enhanced with multi-tenant support and advanced features

-- Tenant and Company Context (for multi-tenancy)
-- Note: These are referenced but not created here as they should exist in core ERP

-- POS Registers
CREATE TABLE IF NOT EXISTS pos_registers (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL, -- Multi-tenant support
    company_id VARCHAR(255), -- Company/Organization
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    location_id INTEGER, -- references locations table
    register_type VARCHAR(50) NOT NULL DEFAULT 'main', -- main, express, self_checkout, mobile
    status VARCHAR(50) NOT NULL DEFAULT 'closed', -- closed, open, suspended, maintenance
    opening_balance DECIMAL(15,2) DEFAULT 0.00,
    current_balance DECIMAL(15,2) DEFAULT 0.00,
    expected_balance DECIMAL(15,2) DEFAULT 0.00,
    opened_at TIMESTAMP,
    opened_by INTEGER, -- references users table
    closed_at TIMESTAMP,
    closed_by INTEGER, -- references users table
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, code),
    CONSTRAINT chk_register_type CHECK (register_type IN ('main', 'express', 'self_checkout', 'mobile')),
    CONSTRAINT chk_register_status CHECK (status IN ('closed', 'open', 'suspended', 'maintenance'))
);

-- Register Shifts (Cashier Shifts)
CREATE TABLE IF NOT EXISTS register_shifts (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    register_id INTEGER NOT NULL REFERENCES pos_registers(id),
    shift_number VARCHAR(50) NOT NULL,
    cashier_id INTEGER NOT NULL, -- references users table
    opening_balance DECIMAL(15,2) NOT NULL,
    closing_balance DECIMAL(15,2),
    expected_balance DECIMAL(15,2),
    variance DECIMAL(15,2) GENERATED ALWAYS AS (closing_balance - expected_balance) STORED,
    total_sales DECIMAL(15,2) DEFAULT 0,
    total_cash_sales DECIMAL(15,2) DEFAULT 0,
    total_card_sales DECIMAL(15,2) DEFAULT 0,
    total_returns DECIMAL(15,2) DEFAULT 0,
    transaction_count INT DEFAULT 0,
    opened_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'open', -- open, closed, reconciled
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, shift_number),
    CONSTRAINT chk_shift_status CHECK (status IN ('open', 'closed', 'reconciled'))
);

-- Register Transactions (Cash In/Out)
CREATE TABLE IF NOT EXISTS register_transactions (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    register_id INTEGER NOT NULL REFERENCES pos_registers(id),
    shift_id INTEGER REFERENCES register_shifts(id),
    transaction_type VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,
    reason VARCHAR(255),
    notes TEXT,
    reference_number VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER, -- references users table
    CONSTRAINT chk_register_txn_type CHECK (transaction_type IN ('opening', 'closing', 'cash_in', 'cash_out', 'cash_drop', 'payout', 'refund_cash'))
);

-- POS Terminals/Devices
CREATE TABLE IF NOT EXISTS pos_terminals (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    location_id INTEGER, -- references locations table
    register_id INTEGER REFERENCES pos_registers(id),
    terminal_code VARCHAR(50) NOT NULL,
    terminal_name VARCHAR(100) NOT NULL,
    device_type VARCHAR(50) NOT NULL, -- desktop, tablet, mobile, kiosk, self_checkout
    hardware_id VARCHAR(255),
    ip_address VARCHAR(45),
    mac_address VARCHAR(17),
    os_version VARCHAR(100),
    app_version VARCHAR(50),
    last_sync_at TIMESTAMP,
    last_heartbeat_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_online BOOLEAN NOT NULL DEFAULT false,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, terminal_code),
    CONSTRAINT chk_device_type CHECK (device_type IN ('desktop', 'tablet', 'mobile', 'kiosk', 'self_checkout'))
);

-- POS Cash Drawers
CREATE TABLE IF NOT EXISTS pos_cash_drawers (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    register_id INTEGER NOT NULL REFERENCES pos_registers(id),
    opening_amount DECIMAL(15,2) DEFAULT 0.00,
    closing_amount DECIMAL(15,2),
    expected_amount DECIMAL(15,2),
    difference_amount DECIMAL(15,2),
    status VARCHAR(20) DEFAULT 'closed', -- open, closed
    opened_by INTEGER, -- references users table
    closed_by INTEGER, -- references users table
    opened_at TIMESTAMP,
    closed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- POS Cash Drawer Transactions
CREATE TABLE IF NOT EXISTS pos_cash_drawer_transactions (
    id SERIAL PRIMARY KEY,
    cash_drawer_id INTEGER NOT NULL REFERENCES pos_cash_drawers(id),
    transaction_type VARCHAR(50) NOT NULL, -- sale, refund, cash_in, cash_out, float
    amount DECIMAL(10,2) NOT NULL,
    reference_type VARCHAR(50), -- transaction_id, etc.
    reference_id INTEGER,
    notes TEXT,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- POS Sessions
CREATE TABLE IF NOT EXISTS pos_sessions (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    register_id INTEGER NOT NULL REFERENCES pos_registers(id),
    cash_drawer_id INTEGER REFERENCES pos_cash_drawers(id),
    shift_id INTEGER REFERENCES register_shifts(id),
    user_id INTEGER NOT NULL, -- references users table
    session_number VARCHAR(50) NOT NULL,
    session_start TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    session_end TIMESTAMP,
    opening_amount DECIMAL(15,2) DEFAULT 0.00,
    closing_amount DECIMAL(15,2),
    total_sales DECIMAL(15,2) DEFAULT 0.00,
    total_refunds DECIMAL(15,2) DEFAULT 0.00,
    total_transactions INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active', -- active, closed
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, session_number)
);

-- POS Transactions
CREATE TABLE IF NOT EXISTS pos_transactions (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    transaction_number VARCHAR(50) NOT NULL,
    session_id INTEGER NOT NULL REFERENCES pos_sessions(id),
    register_id INTEGER NOT NULL REFERENCES pos_registers(id),
    shift_id INTEGER REFERENCES register_shifts(id),
    customer_id INTEGER, -- references customers table
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    transaction_type VARCHAR(50) NOT NULL DEFAULT 'sale', -- sale, return, exchange, void
    status VARCHAR(20) DEFAULT 'completed', -- pending, completed, cancelled, refunded, void
    subtotal DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tip_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) DEFAULT 0.00,
    change_amount DECIMAL(15,2) DEFAULT 0.00,
    cashier_id INTEGER NOT NULL, -- references users table
    manager_id INTEGER, -- for manager approval
    notes TEXT,
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, transaction_number),
    CONSTRAINT chk_transaction_type CHECK (transaction_type IN ('sale', 'return', 'exchange', 'void')),
    CONSTRAINT chk_transaction_status CHECK (status IN ('pending', 'completed', 'cancelled', 'refunded', 'void'))
);

-- POS Transaction Items
CREATE TABLE IF NOT EXISTS pos_transaction_items (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL REFERENCES pos_transactions(id),
    product_id INTEGER NOT NULL, -- references products table
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(15,2) NOT NULL,
    discount_percent DECIMAL(5,2) DEFAULT 0.00,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_rate DECIMAL(5,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    line_total DECIMAL(15,2) GENERATED ALWAYS AS (quantity * unit_price - discount_amount + tax_amount) STORED,
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- POS Payments
CREATE TABLE IF NOT EXISTS pos_payments (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL REFERENCES pos_transactions(id),
    payment_method VARCHAR(50) NOT NULL, -- cash, card, check, gift_card, store_credit, mobile_payment
    amount DECIMAL(15,2) NOT NULL,
    reference_number VARCHAR(100), -- check number, card last 4 digits, transaction ID
    card_type VARCHAR(50), -- visa, mastercard, amex
    status VARCHAR(20) DEFAULT 'completed', -- pending, completed, failed, refunded, void
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_payment_status CHECK (status IN ('pending', 'completed', 'failed', 'refunded', 'void'))
);

-- POS Receipts
CREATE TABLE IF NOT EXISTS pos_receipts (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    transaction_id INTEGER NOT NULL REFERENCES pos_transactions(id),
    receipt_number VARCHAR(50) NOT NULL,
    receipt_type VARCHAR(20) DEFAULT 'sale', -- sale, refund, void
    printed_at TIMESTAMP,
    reprint_count INTEGER DEFAULT 0,
    email_sent BOOLEAN DEFAULT false,
    email_sent_at TIMESTAMP,
    sms_sent BOOLEAN DEFAULT false,
    sms_sent_at TIMESTAMP,
    receipt_data TEXT, -- JSON or HTML receipt data
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, receipt_number)
);

-- Quick Sale Categories
CREATE TABLE IF NOT EXISTS quick_sale_categories (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    location_id INTEGER, -- references locations table
    category_code VARCHAR(50) NOT NULL,
    category_name VARCHAR(100) NOT NULL,
    color_code VARCHAR(20),
    icon VARCHAR(50),
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, category_code)
);

-- Quick Sale Items (Fast Access Buttons)
CREATE TABLE IF NOT EXISTS quick_sale_items (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    location_id INTEGER, -- references locations table
    category_id INTEGER REFERENCES quick_sale_categories(id),
    product_id INTEGER NOT NULL, -- references products table
    button_text VARCHAR(50) NOT NULL,
    button_color VARCHAR(20),
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, product_id, location_id)
);

-- POS Products (POS-specific product settings)
CREATE TABLE IF NOT EXISTS pos_products (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL, -- references products table
    register_id INTEGER REFERENCES pos_registers(id),
    is_available BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    quick_key VARCHAR(10), -- for quick access
    color_code VARCHAR(7), -- hex color for UI
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, register_id)
);

-- POS Customers (POS-specific customer settings)
CREATE TABLE IF NOT EXISTS pos_customers (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL, -- references customers table
    loyalty_points INTEGER DEFAULT 0,
    total_spent DECIMAL(12,2) DEFAULT 0.00,
    last_visit TIMESTAMP,
    is_vip BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Discount Rules (Enhanced)
CREATE TABLE IF NOT EXISTS discount_rules (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    company_id VARCHAR(255), -- references companies table
    rule_code VARCHAR(50) NOT NULL,
    rule_name VARCHAR(100) NOT NULL,
    discount_type VARCHAR(50) NOT NULL, -- percentage, fixed_amount, buy_x_get_y, buy_x_get_discount, bundle
    discount_value DECIMAL(15,4) NOT NULL,
    applies_to VARCHAR(50) NOT NULL, -- all_products, category, specific_products, order_total
    min_purchase_amount DECIMAL(15,2),
    max_discount_amount DECIMAL(15,2),
    buy_quantity INT,
    get_quantity INT,
    customer_group_id INTEGER, -- references customer_groups table
    valid_from TIMESTAMP NOT NULL,
    valid_to TIMESTAMP,
    days_of_week INT[], -- array of days (0=Sunday, 6=Saturday)
    time_from TIME,
    time_to TIME,
    usage_limit INT,
    usage_count INT NOT NULL DEFAULT 0,
    requires_approval BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, rule_code),
    CONSTRAINT chk_discount_type CHECK (discount_type IN ('percentage', 'fixed_amount', 'buy_x_get_y', 'buy_x_get_discount', 'bundle')),
    CONSTRAINT chk_applies_to CHECK (applies_to IN ('all_products', 'category', 'specific_products', 'order_total'))
);

-- Discount Rule Products
CREATE TABLE IF NOT EXISTS discount_rule_products (
    id SERIAL PRIMARY KEY,
    discount_rule_id INTEGER NOT NULL REFERENCES discount_rules(id) ON DELETE CASCADE,
    product_id INTEGER, -- references products table
    category_id INTEGER, -- references product_categories table
    CONSTRAINT chk_product_or_category CHECK (
        (product_id IS NOT NULL AND category_id IS NULL) OR
        (product_id IS NULL AND category_id IS NOT NULL)
    )
);

-- Coupon Codes
CREATE TABLE IF NOT EXISTS coupon_codes (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    discount_rule_id INTEGER REFERENCES discount_rules(id) ON DELETE CASCADE,
    coupon_code VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    max_uses INT,
    max_uses_per_customer INT,
    current_uses INT NOT NULL DEFAULT 0,
    valid_from TIMESTAMP NOT NULL,
    valid_to TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, coupon_code)
);

-- Coupon Usage
CREATE TABLE IF NOT EXISTS coupon_usage (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    coupon_id INTEGER NOT NULL REFERENCES coupon_codes(id) ON DELETE CASCADE,
    transaction_id INTEGER NOT NULL REFERENCES pos_transactions(id) ON DELETE CASCADE,
    customer_id INTEGER, -- references customers table
    discount_amount DECIMAL(15,2) NOT NULL,
    used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Gift Cards
CREATE TABLE IF NOT EXISTS gift_cards (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    company_id VARCHAR(255), -- references companies table
    card_number VARCHAR(50) NOT NULL,
    pin_code VARCHAR(10),
    initial_value DECIMAL(15,2) NOT NULL,
    current_balance DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    purchased_by_customer_id INTEGER, -- references customers table
    recipient_name VARCHAR(255),
    recipient_email VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, used, expired, cancelled, void
    issued_date DATE NOT NULL,
    expiry_date DATE,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, card_number),
    CONSTRAINT chk_gift_card_status CHECK (status IN ('active', 'used', 'expired', 'cancelled', 'void'))
);

-- Gift Card Transactions
CREATE TABLE IF NOT EXISTS gift_card_transactions (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    gift_card_id INTEGER NOT NULL REFERENCES gift_cards(id) ON DELETE CASCADE,
    transaction_type VARCHAR(50) NOT NULL, -- issue, reload, redeem, refund, adjustment, void
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,
    transaction_id INTEGER REFERENCES pos_transactions(id),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER, -- references users table
    CONSTRAINT chk_gift_card_txn_type CHECK (transaction_type IN ('issue', 'reload', 'redeem', 'refund', 'adjustment', 'void'))
);

-- Customer Store Credit
CREATE TABLE IF NOT EXISTS customer_store_credit (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    company_id VARCHAR(255), -- references companies table
    customer_id INTEGER NOT NULL, -- references customers table
    current_balance DECIMAL(15,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, company_id, customer_id)
);

-- Store Credit Transactions
CREATE TABLE IF NOT EXISTS store_credit_transactions (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    credit_id INTEGER NOT NULL REFERENCES customer_store_credit(id) ON DELETE CASCADE,
    transaction_type VARCHAR(50) NOT NULL, -- credit, debit, refund, adjustment, expiry
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,
    transaction_id INTEGER REFERENCES pos_transactions(id),
    reason VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER, -- references users table
    CONSTRAINT chk_credit_txn_type CHECK (transaction_type IN ('credit', 'debit', 'refund', 'adjustment', 'expiry'))
);

-- POS Taxes
CREATE TABLE IF NOT EXISTS pos_taxes (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    rate DECIMAL(5,2) NOT NULL,
    tax_type VARCHAR(20) DEFAULT 'sales', -- sales, service, excise, etc.
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, code)
);

-- POS Shifts (Alias for register_shifts for compatibility)
CREATE TABLE IF NOT EXISTS pos_shifts (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    register_id INTEGER NOT NULL REFERENCES pos_registers(id),
    shift_number VARCHAR(50) NOT NULL,
    user_id INTEGER NOT NULL,
    shift_start TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    shift_end TIMESTAMP,
    opening_amount DECIMAL(15,2) DEFAULT 0.00,
    closing_amount DECIMAL(15,2),
    expected_balance DECIMAL(15,2),
    variance DECIMAL(15,2),
    total_sales DECIMAL(15,2) DEFAULT 0.00,
    total_cash_sales DECIMAL(15,2) DEFAULT 0.00,
    total_card_sales DECIMAL(15,2) DEFAULT 0.00,
    total_returns DECIMAL(15,2) DEFAULT 0.00,
    total_transactions INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active', -- active, closed, reconciled
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, shift_number),
    CONSTRAINT chk_pos_shift_status CHECK (status IN ('active', 'closed', 'reconciled'))
);

-- POS Employees (POS-specific employee settings)
CREATE TABLE IF NOT EXISTS pos_employees (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL, -- references users table
    employee_id VARCHAR(50),
    role VARCHAR(50) DEFAULT 'cashier', -- cashier, manager, supervisor
    permissions TEXT, -- JSON permissions
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
-- Multi-tenant indexes
CREATE INDEX IF NOT EXISTS idx_pos_registers_tenant ON pos_registers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pos_registers_company ON pos_registers(company_id);
CREATE INDEX IF NOT EXISTS idx_pos_sessions_tenant ON pos_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pos_transactions_tenant ON pos_transactions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pos_receipts_tenant ON pos_receipts(tenant_id);

-- Register indexes
CREATE INDEX IF NOT EXISTS idx_pos_sessions_register ON pos_sessions(register_id);
CREATE INDEX IF NOT EXISTS idx_pos_sessions_user ON pos_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_pos_sessions_status ON pos_sessions(status);
CREATE INDEX IF NOT EXISTS idx_pos_sessions_shift ON pos_sessions(shift_id);

-- Transaction indexes
CREATE INDEX IF NOT EXISTS idx_pos_transactions_session ON pos_transactions(session_id);
CREATE INDEX IF NOT EXISTS idx_pos_transactions_date ON pos_transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_pos_transactions_status ON pos_transactions(status);
CREATE INDEX IF NOT EXISTS idx_pos_transactions_customer ON pos_transactions(customer_id);
CREATE INDEX IF NOT EXISTS idx_pos_transactions_register ON pos_transactions(register_id);
CREATE INDEX IF NOT EXISTS idx_pos_transactions_shift ON pos_transactions(shift_id);

-- Payment and receipt indexes
CREATE INDEX IF NOT EXISTS idx_pos_payments_transaction ON pos_payments(transaction_id);
CREATE INDEX IF NOT EXISTS idx_pos_payments_status ON pos_payments(status);
CREATE INDEX IF NOT EXISTS idx_pos_receipts_transaction ON pos_receipts(transaction_id);

-- Cash drawer indexes
CREATE INDEX IF NOT EXISTS idx_pos_cash_drawers_register ON pos_cash_drawers(register_id);
CREATE INDEX IF NOT EXISTS idx_pos_cash_drawer_transactions_drawer ON pos_cash_drawer_transactions(cash_drawer_id);

-- Register shift indexes
CREATE INDEX IF NOT EXISTS idx_register_shifts_register ON register_shifts(register_id);
CREATE INDEX IF NOT EXISTS idx_register_shifts_cashier ON register_shifts(cashier_id);
CREATE INDEX IF NOT EXISTS idx_register_shifts_opened ON register_shifts(opened_at DESC);
CREATE INDEX IF NOT EXISTS idx_register_shifts_status ON register_shifts(status);
CREATE INDEX IF NOT EXISTS idx_register_shifts_tenant ON register_shifts(tenant_id);

-- Register transaction indexes
CREATE INDEX IF NOT EXISTS idx_register_txn_register ON register_transactions(register_id);
CREATE INDEX IF NOT EXISTS idx_register_txn_shift ON register_transactions(shift_id);
CREATE INDEX IF NOT EXISTS idx_register_txn_type ON register_transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_register_txn_created ON register_transactions(created_at DESC);

-- Terminal indexes
CREATE INDEX IF NOT EXISTS idx_pos_terminals_tenant ON pos_terminals(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pos_terminals_location ON pos_terminals(location_id);
CREATE INDEX IF NOT EXISTS idx_pos_terminals_register ON pos_terminals(register_id);
CREATE INDEX IF NOT EXISTS idx_pos_terminals_active ON pos_terminals(is_active);

-- Quick sale indexes
CREATE INDEX IF NOT EXISTS idx_quick_sale_categories_tenant ON quick_sale_categories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_quick_sale_items_tenant ON quick_sale_items(tenant_id);
CREATE INDEX IF NOT EXISTS idx_quick_sale_items_category ON quick_sale_items(category_id);
CREATE INDEX IF NOT EXISTS idx_quick_sale_items_product ON quick_sale_items(product_id);

-- Discount indexes
CREATE INDEX IF NOT EXISTS idx_discount_rules_tenant ON discount_rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_discount_rules_active ON discount_rules(is_active, valid_from, valid_to);
CREATE INDEX IF NOT EXISTS idx_discount_rule_products_rule ON discount_rule_products(discount_rule_id);
CREATE INDEX IF NOT EXISTS idx_coupon_codes_tenant ON coupon_codes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_coupon_codes_code ON coupon_codes(coupon_code);
CREATE INDEX IF NOT EXISTS idx_coupon_usage_coupon ON coupon_usage(coupon_id);
CREATE INDEX IF NOT EXISTS idx_coupon_usage_customer ON coupon_usage(customer_id);

-- Gift card indexes
CREATE INDEX IF NOT EXISTS idx_gift_cards_tenant ON gift_cards(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gift_cards_number ON gift_cards(card_number);
CREATE INDEX IF NOT EXISTS idx_gift_cards_status ON gift_cards(status);
CREATE INDEX IF NOT EXISTS idx_gift_card_txn_card ON gift_card_transactions(gift_card_id);

-- Store credit indexes
CREATE INDEX IF NOT EXISTS idx_customer_store_credit_tenant ON customer_store_credit(tenant_id);
CREATE INDEX IF NOT EXISTS idx_customer_store_credit_customer ON customer_store_credit(customer_id);
CREATE INDEX IF NOT EXISTS idx_store_credit_txn_credit ON store_credit_transactions(credit_id);

-- Tax indexes
CREATE INDEX IF NOT EXISTS idx_pos_taxes_tenant ON pos_taxes(tenant_id);

-- Create triggers for updated_at timestamps
CREATE TRIGGER update_pos_registers_updated_at BEFORE UPDATE ON pos_registers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pos_cash_drawers_updated_at BEFORE UPDATE ON pos_cash_drawers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pos_sessions_updated_at BEFORE UPDATE ON pos_sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pos_transactions_updated_at BEFORE UPDATE ON pos_transactions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pos_products_updated_at BEFORE UPDATE ON pos_products FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pos_customers_updated_at BEFORE UPDATE ON pos_customers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pos_discounts_updated_at BEFORE UPDATE ON pos_discounts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pos_taxes_updated_at BEFORE UPDATE ON pos_taxes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pos_shifts_updated_at BEFORE UPDATE ON pos_shifts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pos_employees_updated_at BEFORE UPDATE ON pos_employees FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pos_terminals_updated_at BEFORE UPDATE ON pos_terminals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_quick_sale_items_updated_at BEFORE UPDATE ON quick_sale_items FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_discount_rules_updated_at BEFORE UPDATE ON discount_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_coupon_codes_updated_at BEFORE UPDATE ON coupon_codes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_gift_cards_updated_at BEFORE UPDATE ON gift_cards FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_customer_store_credit_updated_at BEFORE UPDATE ON customer_store_credit FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
