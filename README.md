# ERP POS Module

A comprehensive point of sale (POS) system for the LinearBits ERP system.

## Features

- ✅ Touchscreen interface for retail
- ✅ Payment processing integration
- ✅ Receipt printing and management
- ✅ Cash drawer and register management
- ✅ Customer management with loyalty programs
- ✅ Inventory integration
- ✅ Sales analytics and reporting
- ✅ Multi-tenant support
- ✅ Advanced features: Gift cards, Coupons, Store credit

## Installation

This module can be installed through the LinearBits ERP Marketplace or directly from GitHub.

## Implementation Status

### Database Schema ✅
Enhanced schema with:
- Multi-tenant support (`tenant_id` on all tables)
- Advanced features (gift cards, coupons, loyalty programs)
- Register shifts and reconciliation
- Quick sale items for fast checkout
- Comprehensive indexes for performance

### Handlers Implemented ✅

#### Core Handlers
- `pos_handler.go` - Sessions, transactions, registers, receipts, analytics
- `product_handler.go` - POS products and quick sale items
- `gift_card_handler.go` - Gift card issuance, redemption, balance tracking
- `discount_handler.go` - Discount rules and coupon management
- `tax_handler.go` - Tax rate management
- `shift_handler.go` - Cashier shift management with reconciliation
- `customer_handler.go` - Customer loyalty operations

#### Domain Models ✅
- Complete type definitions for all POS entities
- Metadata support for extensibility
- Request/Response DTOs

## Usage

Once installed, the POS module will be available in your ERP navigation menu under "POS".

## API Endpoints

### Sessions
- `GET /api/v1/pos/sessions` - List POS sessions
- `POST /api/v1/pos/sessions` - Start new session
- `POST /api/v1/pos/sessions/{id}/close` - Close session

### Transactions
- `GET /api/v1/pos/transactions` - List transactions
- `POST /api/v1/pos/transactions` - Create transaction
- `GET /api/v1/pos/transactions/{id}` - Get transaction

### Products
- `GET /api/v1/pos/products` - List POS products
- `POST /api/v1/pos/products` - Link product to POS
- `GET /api/v1/pos/quick-sale/items` - List quick sale items
- `POST /api/v1/pos/quick-sale/items` - Add quick sale item

### Gift Cards
- `GET /api/v1/pos/gift-cards` - List gift cards
- `POST /api/v1/pos/gift-cards` - Issue gift card
- `POST /api/v1/pos/gift-cards/redeem` - Redeem gift card
- `GET /api/v1/pos/gift-cards/number/{number}` - Check balance

### Discounts & Coupons
- `GET /api/v1/pos/discounts` - List discount rules
- `POST /api/v1/pos/discounts` - Create discount rule
- `POST /api/v1/pos/coupons/validate` - Validate coupon
- `POST /api/v1/pos/coupons` - Create coupon code

### Shifts
- `GET /api/v1/pos/shifts` - List shifts
- `POST /api/v1/pos/shifts` - Start shift
- `POST /api/v1/pos/shifts/{id}/close` - Close shift

### Customers
- `GET /api/v1/pos/customers` - List customers with loyalty info
- `GET /api/v1/pos/customers/loyalty` - Get loyalty information
- `POST /api/v1/pos/customers/loyalty` - Update loyalty points

### Taxes
- `GET /api/v1/pos/taxes` - List tax rates
- `POST /api/v1/pos/taxes` - Create tax rate

## Permissions

- `pos.transactions.view` - View transactions
- `pos.transactions.create` - Create transactions
- `pos.transactions.edit` - Edit transactions
- `pos.sessions.view` - View sessions
- `pos.sessions.create` - Create sessions
- `pos.sessions.close` - Close sessions
- `pos.payments.view` - View payments
- `pos.payments.create` - Process payments
- `pos.receipts.view` - View receipts
- `pos.receipts.print` - Print receipts
- `pos.registers.view` - View registers
- `pos.registers.create` - Create registers
- `pos.giftcards.view` - View gift cards
- `pos.giftcards.create` - Issue gift cards
- `pos.discounts.view` - View discounts
- `pos.discounts.create` - Create discounts
- `pos.shifts.view` - View shifts
- `pos.shifts.create` - Start shifts
- `pos.shifts.close` - Close shifts

## Database Tables

This module uses the following database tables:

### Core Tables
- `pos_registers` - Register definitions
- `pos_sessions` - POS session records
- `pos_transactions` - Transaction headers
- `pos_transaction_items` - Transaction line items
- `pos_payments` - Payment records
- `pos_receipts` - Receipt records
- `pos_cash_drawers` - Cash drawer records

### Advanced Features
- `register_shifts` - Cashier shift tracking
- `register_transactions` - Cash in/out operations
- `pos_terminals` - Device management
- `quick_sale_categories` - Quick sale categories
- `quick_sale_items` - Fast checkout products
- `discount_rules` - Discount management
- `discount_rule_products` - Product/category linking
- `coupon_codes` - Coupon codes
- `coupon_usage` - Coupon usage tracking
- `gift_cards` - Gift card management
- `gift_card_transactions` - Gift card transactions
- `customer_store_credit` - Store credit balances
- `store_credit_transactions` - Credit transactions
- `pos_taxes` - Tax rates
- `pos_products` - POS-specific product settings
- `pos_customers` - POS-specific customer settings

## Multi-Tenant Support

All handlers include multi-tenant context extraction and filtering:
- Tenant ID is extracted from request headers (`X-Tenant-ID`)
- All queries are automatically filtered by `tenant_id`
- Secure data isolation between tenants

## Advanced Features

### Gift Cards
- Issue gift cards with configurable amounts
- Support for expiry dates
- Track gift card balance and usage
- Redeem gift cards for payments

### Coupons & Discounts
- Support for percentage and fixed discounts
- Buy-X-Get-Y promotions
- Time-based discounts
- Usage limits and validation
- Automatic discount calculation

### Loyalty Program
- Track customer loyalty points
- Points earned per transaction
- Points redemption for discounts
- Customer lifetime value tracking

### Register Shifts
- Cashier shift tracking
- Opening and closing balance tracking
- Automatic variance calculation
- Shift reconciliation reports

### Quick Sale Items
- Fast checkout for popular products
- Categorized quick sale buttons
- Customizable display order
- Product availability toggling

## Database Migration

To apply the schema changes:

```bash
psql -U your_user -d your_database -f migrations/001_create_pos_tables.sql
```

## Testing

Test all endpoints with real data:

```bash
# Test session creation
curl -X POST http://localhost:8080/api/v1/pos/sessions \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: your_tenant_id" \
  -d '{
    "register_id": 1,
    "opening_amount": 100.00
  }'

# Test gift card issuance
curl -X POST http://localhost:8080/api/v1/pos/gift-cards \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: your_tenant_id" \
  -d '{
    "initial_value": 50.00,
    "currency": "USD"
  }'
```

## License

MIT License - see LICENSE file for details.

## Support

For support and questions, please open an issue on GitHub or contact the LinearBits team.
