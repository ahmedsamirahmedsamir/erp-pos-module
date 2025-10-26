package main

import "time"

// POSTransaction represents a POS transaction
type POSTransaction struct {
	ID                int       `json:"id"`
	TransactionNumber string    `json:"transaction_number"`
	ShiftID           *int      `json:"shift_id"`
	CustomerID        *int      `json:"customer_id"`
	Subtotal          float64   `json:"subtotal"`
	TaxAmount         float64   `json:"tax_amount"`
	DiscountAmount    float64   `json:"discount_amount"`
	TotalAmount       float64   `json:"total_amount"`
	PaymentMethod     string    `json:"payment_method"`
	Status            string    `json:"status"`
	CreatedBy         int       `json:"created_by"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// POSShift represents a cashier shift
type POSShift struct {
	ID           int        `json:"id"`
	ShiftNumber  string     `json:"shift_number"`
	CashierID    int        `json:"cashier_id"`
	RegisterID   int        `json:"register_id"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	OpeningCash  float64    `json:"opening_cash"`
	ClosingCash  *float64   `json:"closing_cash"`
	ExpectedCash *float64   `json:"expected_cash"`
	Variance     *float64   `json:"variance"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
