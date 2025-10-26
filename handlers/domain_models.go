package pos

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Multi-tenant context helpers
const (
	ContextKeyTenantID = "tenant_id"
	ContextKeyUserID   = "user_id"
)

// =================================================================
// CORE ENTITIES
// =================================================================

// POSRegister represents a cash register/till
type POSRegister struct {
	ID              int        `json:"id" db:"id"`
	TenantID        string     `json:"tenant_id" db:"tenant_id"`
	CompanyID       *string    `json:"company_id" db:"company_id"`
	Name            string     `json:"name" db:"name"`
	Code            string     `json:"code" db:"code"`
	LocationID      *int       `json:"location_id" db:"location_id"`
	RegisterType    string     `json:"register_type" db:"register_type"` // main, express, self_checkout, mobile
	Status          string     `json:"status" db:"status"`               // closed, open, suspended, maintenance
	OpeningBalance  float64    `json:"opening_balance" db:"opening_balance"`
	CurrentBalance  float64    `json:"current_balance" db:"current_balance"`
	ExpectedBalance float64    `json:"expected_balance" db:"expected_balance"`
	OpenedAt        *time.Time `json:"opened_at" db:"opened_at"`
	OpenedBy        *int       `json:"opened_by" db:"opened_by"`
	ClosedAt        *time.Time `json:"closed_at" db:"closed_at"`
	ClosedBy        *int       `json:"closed_by" db:"closed_by"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	Metadata        Metadata   `json:"metadata" db:"metadata"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// RegisterShift represents a cashier shift
type RegisterShift struct {
	ID               int        `json:"id" db:"id"`
	TenantID         string     `json:"tenant_id" db:"tenant_id"`
	RegisterID       int        `json:"register_id" db:"register_id"`
	ShiftNumber      string     `json:"shift_number" db:"shift_number"`
	CashierID        int        `json:"cashier_id" db:"cashier_id"`
	OpeningBalance   float64    `json:"opening_balance" db:"opening_balance"`
	ClosingBalance   *float64   `json:"closing_balance" db:"closing_balance"`
	ExpectedBalance  *float64   `json:"expected_balance" db:"expected_balance"`
	Variance         *float64   `json:"variance" db:"variance"`
	TotalSales       float64    `json:"total_sales" db:"total_sales"`
	TotalCashSales   float64    `json:"total_cash_sales" db:"total_cash_sales"`
	TotalCardSales   float64    `json:"total_card_sales" db:"total_card_sales"`
	TotalReturns     float64    `json:"total_returns" db:"total_returns"`
	TransactionCount int        `json:"transaction_count" db:"transaction_count"`
	OpenedAt         time.Time  `json:"opened_at" db:"opened_at"`
	ClosedAt         *time.Time `json:"closed_at" db:"closed_at"`
	Status           string     `json:"status" db:"status"` // open, closed, reconciled
	Notes            *string    `json:"notes" db:"notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// POSTerminal represents a POS device/terminal
type POSTerminal struct {
	ID              int        `json:"id" db:"id"`
	TenantID        string     `json:"tenant_id" db:"tenant_id"`
	LocationID      *int       `json:"location_id" db:"location_id"`
	RegisterID      *int       `json:"register_id" db:"register_id"`
	TerminalCode    string     `json:"terminal_code" db:"terminal_code"`
	TerminalName    string     `json:"terminal_name" db:"terminal_name"`
	DeviceType      string     `json:"device_type" db:"device_type"` // desktop, tablet, mobile, kiosk, self_checkout
	HardwareID      *string    `json:"hardware_id" db:"hardware_id"`
	IPAddress       *string    `json:"ip_address" db:"ip_address"`
	MACAddress      *string    `json:"mac_address" db:"mac_address"`
	OSVersion       *string    `json:"os_version" db:"os_version"`
	AppVersion      *string    `json:"app_version" db:"app_version"`
	LastSyncAt      *time.Time `json:"last_sync_at" db:"last_sync_at"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at" db:"last_heartbeat_at"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	IsOnline        bool       `json:"is_online" db:"is_online"`
	Settings        Metadata   `json:"settings" db:"settings"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// POSSession represents a POS session
type POSSession struct {
	ID                int          `json:"id" db:"id"`
	TenantID          string       `json:"tenant_id" db:"tenant_id"`
	RegisterID        int          `json:"register_id" db:"register_id"`
	CashDrawerID      *int         `json:"cash_drawer_id" db:"cash_drawer_id"`
	ShiftID           *int         `json:"shift_id" db:"shift_id"`
	UserID            int          `json:"user_id" db:"user_id"`
	SessionNumber     string       `json:"session_number" db:"session_number"`
	SessionStart      time.Time    `json:"session_start" db:"session_start"`
	SessionEnd        *time.Time   `json:"session_end" db:"session_end"`
	OpeningAmount     float64      `json:"opening_amount" db:"opening_amount"`
	ClosingAmount     *float64     `json:"closing_amount" db:"closing_amount"`
	TotalSales        float64      `json:"total_sales" db:"total_sales"`
	TotalRefunds      float64      `json:"total_refunds" db:"total_refunds"`
	TotalTransactions int          `json:"total_transactions" db:"total_transactions"`
	Status            string       `json:"status" db:"status"` // active, closed
	Notes             *string      `json:"notes" db:"notes"`
	Metadata          Metadata     `json:"metadata" db:"metadata"`
	CreatedAt         time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at" db:"updated_at"`
	Register          *POSRegister `json:"register,omitempty"`
	User              *User        `json:"user,omitempty"`
}

// POSTransaction represents a sales transaction
type POSTransaction struct {
	ID                int                  `json:"id" db:"id"`
	TenantID          string               `json:"tenant_id" db:"tenant_id"`
	TransactionNumber string               `json:"transaction_number" db:"transaction_number"`
	SessionID         int                  `json:"session_id" db:"session_id"`
	RegisterID        int                  `json:"register_id" db:"register_id"`
	ShiftID           *int                 `json:"shift_id" db:"shift_id"`
	CustomerID        *int                 `json:"customer_id" db:"customer_id"`
	TransactionDate   time.Time            `json:"transaction_date" db:"transaction_date"`
	TransactionType   string               `json:"transaction_type" db:"transaction_type"` // sale, return, exchange, void
	Status            string               `json:"status" db:"status"`                     // pending, completed, cancelled, refunded, void
	Subtotal          float64              `json:"subtotal" db:"subtotal"`
	TaxAmount         float64              `json:"tax_amount" db:"tax_amount"`
	DiscountAmount    float64              `json:"discount_amount" db:"discount_amount"`
	TipAmount         float64              `json:"tip_amount" db:"tip_amount"`
	TotalAmount       float64              `json:"total_amount" db:"total_amount"`
	ChangeAmount      float64              `json:"change_amount" db:"change_amount"`
	CashierID         int                  `json:"cashier_id" db:"cashier_id"`
	ManagerID         *int                 `json:"manager_id" db:"manager_id"`
	Notes             *string              `json:"notes" db:"notes"`
	CustomFields      Metadata             `json:"custom_fields" db:"custom_fields"`
	CreatedAt         time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at" db:"updated_at"`
	Items             []POSTransactionItem `json:"items,omitempty"`
	Payments          []POSPayment         `json:"payments,omitempty"`
	Customer          *Customer            `json:"customer,omitempty"`
	Cashier           *User                `json:"cashier,omitempty"`
	Register          *POSRegister         `json:"register,omitempty"`
	Session           *POSSession          `json:"session,omitempty"`
}

// POSTransactionItem represents a line item in a transaction
type POSTransactionItem struct {
	ID              int       `json:"id" db:"id"`
	TransactionID   int       `json:"transaction_id" db:"transaction_id"`
	ProductID       int       `json:"product_id" db:"product_id"`
	Quantity        int       `json:"quantity" db:"quantity"`
	UnitPrice       float64   `json:"unit_price" db:"unit_price"`
	DiscountPercent float64   `json:"discount_percent" db:"discount_percent"`
	DiscountAmount  float64   `json:"discount_amount" db:"discount_amount"`
	TaxRate         float64   `json:"tax_rate" db:"tax_rate"`
	TaxAmount       float64   `json:"tax_amount" db:"tax_amount"`
	LineTotal       float64   `json:"line_total" db:"line_total"`
	Notes           *string   `json:"notes" db:"notes"`
	Metadata        Metadata  `json:"metadata" db:"metadata"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	Product         *Product  `json:"product,omitempty"`
}

// POSPayment represents a payment for a transaction
type POSPayment struct {
	ID              int       `json:"id" db:"id"`
	TransactionID   int       `json:"transaction_id" db:"transaction_id"`
	PaymentMethod   string    `json:"payment_method" db:"payment_method"`
	Amount          float64   `json:"amount" db:"amount"`
	ReferenceNumber *string   `json:"reference_number" db:"reference_number"`
	CardType        *string   `json:"card_type" db:"card_type"`
	Status          string    `json:"status" db:"status"`
	ProcessedAt     time.Time `json:"processed_at" db:"processed_at"`
	Notes           *string   `json:"notes" db:"notes"`
	Metadata        Metadata  `json:"metadata" db:"metadata"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// POSReceipt represents a receipt
type POSReceipt struct {
	ID            int        `json:"id" db:"id"`
	TenantID      string     `json:"tenant_id" db:"tenant_id"`
	TransactionID int        `json:"transaction_id" db:"transaction_id"`
	ReceiptNumber string     `json:"receipt_number" db:"receipt_number"`
	ReceiptType   string     `json:"receipt_type" db:"receipt_type"` // sale, refund, void
	PrintedAt     *time.Time `json:"printed_at" db:"printed_at"`
	ReprintCount  int        `json:"reprint_count" db:"reprint_count"`
	EmailSent     bool       `json:"email_sent" db:"email_sent"`
	EmailSentAt   *time.Time `json:"email_sent_at" db:"email_sent_at"`
	SMSSent       bool       `json:"sms_sent" db:"sms_sent"`
	SMSSentAt     *time.Time `json:"sms_sent_at" db:"sms_sent_at"`
	ReceiptData   string     `json:"receipt_data" db:"receipt_data"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// QuickSaleCategory represents a category for quick sale items
type QuickSaleCategory struct {
	ID           int       `json:"id" db:"id"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	LocationID   *int      `json:"location_id" db:"location_id"`
	CategoryCode string    `json:"category_code" db:"category_code"`
	CategoryName string    `json:"category_name" db:"category_name"`
	ColorCode    *string   `json:"color_code" db:"color_code"`
	Icon         *string   `json:"icon" db:"icon"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// QuickSaleItem represents a quick access button for a product
type QuickSaleItem struct {
	ID           int                `json:"id" db:"id"`
	TenantID     string             `json:"tenant_id" db:"tenant_id"`
	LocationID   *int               `json:"location_id" db:"location_id"`
	CategoryID   *int               `json:"category_id" db:"category_id"`
	ProductID    int                `json:"product_id" db:"product_id"`
	ButtonText   string             `json:"button_text" db:"button_text"`
	ButtonColor  *string            `json:"button_color" db:"button_color"`
	DisplayOrder int                `json:"display_order" db:"display_order"`
	IsActive     bool               `json:"is_active" db:"is_active"`
	CreatedAt    time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" db:"updated_at"`
	Product      *Product           `json:"product,omitempty"`
	Category     *QuickSaleCategory `json:"category,omitempty"`
}

// POSProduct represents POS-specific product settings
type POSProduct struct {
	ID           int       `json:"id" db:"id"`
	ProductID    int       `json:"product_id" db:"product_id"`
	RegisterID   *int      `json:"register_id" db:"register_id"`
	IsAvailable  bool      `json:"is_available" db:"is_available"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	QuickKey     *string   `json:"quick_key" db:"quick_key"`
	ColorCode    *string   `json:"color_code" db:"color_code"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	Product      *Product  `json:"product,omitempty"`
}

// POSCustomer represents POS-specific customer settings
type POSCustomer struct {
	ID            int        `json:"id" db:"id"`
	CustomerID    int        `json:"customer_id" db:"customer_id"`
	LoyaltyPoints int        `json:"loyalty_points" db:"loyalty_points"`
	TotalSpent    float64    `json:"total_spent" db:"total_spent"`
	LastVisit     *time.Time `json:"last_visit" db:"last_visit"`
	IsVIP         bool       `json:"is_vip" db:"is_vip"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// DiscountRule represents a discount or promotion rule
type DiscountRule struct {
	ID                int        `json:"id" db:"id"`
	TenantID          string     `json:"tenant_id" db:"tenant_id"`
	CompanyID         *string    `json:"company_id" db:"company_id"`
	RuleCode          string     `json:"rule_code" db:"rule_code"`
	RuleName          string     `json:"rule_name" db:"rule_name"`
	DiscountType      string     `json:"discount_type" db:"discount_type"` // percentage, fixed_amount, buy_x_get_y, buy_x_get_discount, bundle
	DiscountValue     float64    `json:"discount_value" db:"discount_value"`
	AppliesTo         string     `json:"applies_to" db:"applies_to"` // all_products, category, specific_products, order_total
	MinPurchaseAmount *float64   `json:"min_purchase_amount" db:"min_purchase_amount"`
	MaxDiscountAmount *float64   `json:"max_discount_amount" db:"max_discount_amount"`
	BuyQuantity       *int       `json:"buy_quantity" db:"buy_quantity"`
	GetQuantity       *int       `json:"get_quantity" db:"get_quantity"`
	CustomerGroupID   *int       `json:"customer_group_id" db:"customer_group_id"`
	ValidFrom         time.Time  `json:"valid_from" db:"valid_from"`
	ValidTo           *time.Time `json:"valid_to" db:"valid_to"`
	DaysOfWeek        []int      `json:"days_of_week" db:"days_of_week"`
	TimeFrom          *string    `json:"time_from" db:"time_from"`
	TimeTo            *string    `json:"time_to" db:"time_to"`
	UsageLimit        *int       `json:"usage_limit" db:"usage_limit"`
	UsageCount        int        `json:"usage_count" db:"usage_count"`
	RequiresApproval  bool       `json:"requires_approval" db:"requires_approval"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	Priority          int        `json:"priority" db:"priority"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	Products          []*Product `json:"products,omitempty"`
}

// CouponCode represents a redeemable coupon
type CouponCode struct {
	ID                 int           `json:"id" db:"id"`
	TenantID           string        `json:"tenant_id" db:"tenant_id"`
	DiscountRuleID     *int          `json:"discount_rule_id" db:"discount_rule_id"`
	CouponCode         string        `json:"coupon_code" db:"coupon_code"`
	Description        *string       `json:"description" db:"description"`
	MaxUses            *int          `json:"max_uses" db:"max_uses"`
	MaxUsesPerCustomer *int          `json:"max_uses_per_customer" db:"max_uses_per_customer"`
	CurrentUses        int           `json:"current_uses" db:"current_uses"`
	ValidFrom          time.Time     `json:"valid_from" db:"valid_from"`
	ValidTo            *time.Time    `json:"valid_to" db:"valid_to"`
	IsActive           bool          `json:"is_active" db:"is_active"`
	CreatedAt          time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at" db:"updated_at"`
	Rule               *DiscountRule `json:"rule,omitempty"`
}

// CouponUsage tracks coupon usage
type CouponUsage struct {
	ID             int       `json:"id" db:"id"`
	TenantID       string    `json:"tenant_id" db:"tenant_id"`
	CouponID       int       `json:"coupon_id" db:"coupon_id"`
	TransactionID  int       `json:"transaction_id" db:"transaction_id"`
	CustomerID     *int      `json:"customer_id" db:"customer_id"`
	DiscountAmount float64   `json:"discount_amount" db:"discount_amount"`
	UsedAt         time.Time `json:"used_at" db:"used_at"`
}

// POSTax represents a tax rate
type POSTax struct {
	ID          int       `json:"id" db:"id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`
	Rate        float64   `json:"rate" db:"rate"`
	TaxType     string    `json:"tax_type" db:"tax_type"`
	Description *string   `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// GiftCard represents a gift card
type GiftCard struct {
	ID                    int        `json:"id" db:"id"`
	TenantID              string     `json:"tenant_id" db:"tenant_id"`
	CompanyID             *string    `json:"company_id" db:"company_id"`
	CardNumber            string     `json:"card_number" db:"card_number"`
	PINCode               *string    `json:"pin_code" db:"pin_code"`
	InitialValue          float64    `json:"initial_value" db:"initial_value"`
	CurrentBalance        float64    `json:"current_balance" db:"current_balance"`
	Currency              string     `json:"currency" db:"currency"`
	PurchasedByCustomerID *int       `json:"purchased_by_customer_id" db:"purchased_by_customer_id"`
	RecipientName         *string    `json:"recipient_name" db:"recipient_name"`
	RecipientEmail        *string    `json:"recipient_email" db:"recipient_email"`
	Status                string     `json:"status" db:"status"`
	IssuedDate            time.Time  `json:"issued_date" db:"issued_date"`
	ExpiryDate            *time.Time `json:"expiry_date" db:"expiry_date"`
	LastUsedAt            *time.Time `json:"last_used_at" db:"last_used_at"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// GiftCardTransaction represents a gift card transaction
type GiftCardTransaction struct {
	ID              int       `json:"id" db:"id"`
	TenantID        string    `json:"tenant_id" db:"tenant_id"`
	GiftCardID      int       `json:"gift_card_id" db:"gift_card_id"`
	TransactionType string    `json:"transaction_type" db:"transaction_type"`
	Amount          float64   `json:"amount" db:"amount"`
	BalanceBefore   float64   `json:"balance_before" db:"balance_before"`
	BalanceAfter    float64   `json:"balance_after" db:"balance_after"`
	TransactionID   *int      `json:"transaction_id" db:"transaction_id"`
	Notes           *string   `json:"notes" db:"notes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	CreatedBy       *int      `json:"created_by" db:"created_by"`
}

// CustomerStoreCredit represents customer store credit balance
type CustomerStoreCredit struct {
	ID             int       `json:"id" db:"id"`
	TenantID       string    `json:"tenant_id" db:"tenant_id"`
	CompanyID      *string   `json:"company_id" db:"company_id"`
	CustomerID     int       `json:"customer_id" db:"customer_id"`
	CurrentBalance float64   `json:"current_balance" db:"current_balance"`
	Currency       string    `json:"currency" db:"currency"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	Customer       *Customer `json:"customer,omitempty"`
}

// StoreCreditTransaction represents a store credit transaction
type StoreCreditTransaction struct {
	ID              int       `json:"id" db:"id"`
	TenantID        string    `json:"tenant_id" db:"tenant_id"`
	CreditID        int       `json:"credit_id" db:"credit_id"`
	TransactionType string    `json:"transaction_type" db:"transaction_type"`
	Amount          float64   `json:"amount" db:"amount"`
	BalanceBefore   float64   `json:"balance_before" db:"balance_before"`
	BalanceAfter    float64   `json:"balance_after" db:"balance_after"`
	TransactionID   *int      `json:"transaction_id" db:"transaction_id"`
	Reason          *string   `json:"reason" db:"reason"`
	Notes           *string   `json:"notes" db:"notes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	CreatedBy       *int      `json:"created_by" db:"created_by"`
}

// =================================================================
// RELATIONSHIP ENTITIES (references to core ERP)
// =================================================================

// User represents a system user (cashier, manager, etc.)
type User struct {
	ID        int     `json:"id"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     string  `json:"email"`
	Role      string  `json:"role"`
}

// Customer represents a customer
type Customer struct {
	ID        int     `json:"id"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
}

// Product represents a product
type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	SKU      string  `json:"sku"`
	Price    float64 `json:"price"`
	Barcode  *string `json:"barcode"`
	ImageURL *string `json:"image_url"`
}

// =================================================================
// UTILITY TYPES
// =================================================================

// Metadata is a JSONB field that can store arbitrary JSON data
type Metadata map[string]interface{}

// Value implements the driver.Valuer interface
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte("{}"), m)
	}
	return json.Unmarshal(bytes, m)
}

// DayOfWeek represents days of the week
type DayOfWeek int

const (
	Sunday DayOfWeek = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// =================================================================
// REQUEST/RESPONSE DTOs
// =================================================================

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// PaginationResponse represents paginated response metadata
type PaginationResponse struct {
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	TotalCount  int  `json:"total_count"`
	Limit       int  `json:"limit"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}

// FilterRequest represents common filtering parameters
type FilterRequest struct {
	Search   string     `json:"search"`
	Status   string     `json:"status"`
	DateFrom *time.Time `json:"date_from"`
	DateTo   *time.Time `json:"date_to"`
}

// SortRequest represents sorting parameters
type SortRequest struct {
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"` // asc, desc
}
