package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	sdk "github.com/linearbits/erp-backend/pkg/module-sdk"
	"go.uber.org/zap"
)

// POSHandler handles POS operations
type POSHandler struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewPOSHandler creates a new POS handler
func NewPOSHandler(db *sqlx.DB, logger *zap.Logger) *POSHandler {
	return &POSHandler{db: db, logger: logger}
}

// GetTransactions retrieves POS transactions
func (h *POSHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	shiftID := r.URL.Query().Get("shift_id")
	status := r.URL.Query().Get("status")

	qb := sdk.NewQueryBuilder("SELECT * FROM pos_transactions WHERE 1=1")
	qb.AddOptionalCondition("shift_id = $%d", shiftID)
	qb.AddOptionalCondition("status = $%d", status)

	query, args := qb.Build()
	query += " ORDER BY created_at DESC LIMIT 100"

	var transactions []POSTransaction
	if err := h.db.Select(&transactions, query, args...); err != nil {
		h.logger.Error("Failed to fetch transactions", zap.Error(err))
		sdk.WriteInternalError(w, "Failed to fetch transactions")
		return
	}

	sdk.WriteSuccess(w, map[string]interface{}{
		"transactions": transactions,
		"count":        len(transactions),
	})
}

// CreateTransaction creates a new POS transaction
func (h *POSHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req POSTransaction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sdk.WriteBadRequest(w, "Invalid request body")
		return
	}

	txnNumber := fmt.Sprintf("POS-%d", time.Now().Unix())

	query := `
		INSERT INTO pos_transactions (transaction_number, shift_id, customer_id, subtotal, 
		                             tax_amount, discount_amount, total_amount, payment_method, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1)
		RETURNING id, created_at
	`

	var id int
	var createdAt time.Time

	err := h.db.QueryRow(query, txnNumber, req.ShiftID, req.CustomerID, req.Subtotal,
		req.TaxAmount, req.DiscountAmount, req.TotalAmount, req.PaymentMethod).Scan(&id, &createdAt)

	if err != nil {
		h.logger.Error("Failed to create transaction", zap.Error(err))
		sdk.WriteInternalError(w, "Failed to create transaction")
		return
	}

	sdk.WriteCreated(w, map[string]interface{}{
		"id":                 id,
		"transaction_number": txnNumber,
		"created_at":         createdAt,
		"message":            "Transaction created successfully",
	})
}

// GetShifts retrieves POS shifts
func (h *POSHandler) GetShifts(w http.ResponseWriter, r *http.Request) {
	cashierID := r.URL.Query().Get("cashier_id")
	status := r.URL.Query().Get("status")

	qb := sdk.NewQueryBuilder("SELECT * FROM pos_shifts WHERE 1=1")
	qb.AddOptionalCondition("cashier_id = $%d", cashierID)
	qb.AddOptionalCondition("status = $%d", status)

	query, args := qb.Build()
	query += " ORDER BY start_time DESC LIMIT 50"

	var shifts []POSShift
	if err := h.db.Select(&shifts, query, args...); err != nil {
		h.logger.Error("Failed to fetch shifts", zap.Error(err))
		sdk.WriteInternalError(w, "Failed to fetch shifts")
		return
	}

	sdk.WriteSuccess(w, map[string]interface{}{
		"shifts": shifts,
		"count":  len(shifts),
	})
}

// CreateShift creates a new shift
func (h *POSHandler) CreateShift(w http.ResponseWriter, r *http.Request) {
	var req POSShift
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sdk.WriteBadRequest(w, "Invalid request body")
		return
	}

	shiftNumber := fmt.Sprintf("SH-%d", time.Now().Unix())

	query := `
		INSERT INTO pos_shifts (shift_number, cashier_id, register_id, start_time, opening_cash, status)
		VALUES ($1, $2, $3, NOW(), $4, 'open')
		RETURNING id, created_at
	`

	var id int
	var createdAt time.Time

	err := h.db.QueryRow(query, shiftNumber, req.CashierID, req.RegisterID, req.OpeningCash).
		Scan(&id, &createdAt)

	if err != nil {
		h.logger.Error("Failed to create shift", zap.Error(err))
		sdk.WriteInternalError(w, "Failed to create shift")
		return
	}

	sdk.WriteCreated(w, map[string]interface{}{
		"id":           id,
		"shift_number": shiftNumber,
		"created_at":   createdAt,
		"message":      "Shift created successfully",
	})
}
