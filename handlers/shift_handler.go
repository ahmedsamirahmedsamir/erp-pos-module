package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// ShiftHandler handles shift-related operations
type ShiftHandler struct {
	db          *sqlx.DB
	logger      *zap.Logger
	baseHandler *POSHandler
}

// NewShiftHandler creates a new shift handler
func NewShiftHandler(db *sqlx.DB, logger *zap.Logger) *ShiftHandler {
	return &ShiftHandler{
		db:          db,
		logger:      logger,
		baseHandler: NewPOSHandler(db, logger),
	}
}

// GetPOSShifts retrieves all shifts
func (h *ShiftHandler) GetPOSShifts(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	registerID := r.URL.Query().Get("register_id")
	status := r.URL.Query().Get("status")

	query := `SELECT * FROM register_shifts WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argIndex := 2

	if registerID != "" {
		query += fmt.Sprintf(" AND register_id = $%d", argIndex)
		args = append(args, registerID)
		argIndex++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += " ORDER BY opened_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch shifts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var shifts []RegisterShift
	for rows.Next() {
		var shift RegisterShift
		err := rows.Scan(&shift.ID, &shift.TenantID, &shift.RegisterID, &shift.ShiftNumber,
			&shift.CashierID, &shift.OpeningBalance, &shift.ClosingBalance, &shift.ExpectedBalance,
			&shift.Variance, &shift.TotalSales, &shift.TotalCashSales, &shift.TotalCardSales,
			&shift.TotalReturns, &shift.TransactionCount, &shift.OpenedAt, &shift.ClosedAt,
			&shift.Status, &shift.Notes, &shift.CreatedAt)
		if err != nil {
			continue
		}
		shifts = append(shifts, shift)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"shifts": shifts,
		"count":  len(shifts),
	})
}

// CreatePOSShift starts a new shift
func (h *ShiftHandler) CreatePOSShift(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		RegisterID     int     `json:"register_id" validate:"required"`
		OpeningBalance float64 `json:"opening_balance" validate:"required"`
		Notes          *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, _ := h.baseHandler.getUserID(r)

	// Generate shift number
	shiftNumber := fmt.Sprintf("SHIFT-%d", time.Now().Unix())

	query := `
		INSERT INTO register_shifts (tenant_id, register_id, shift_number, cashier_id, opening_balance, status)
		VALUES ($1, $2, $3, $4, $5, 'open')
		RETURNING id, created_at
	`

	var id int
	var createdAt time.Time

	err = h.db.QueryRow(query, tenantID, req.RegisterID, shiftNumber, userID, req.OpeningBalance).
		Scan(&id, &createdAt)
	if err != nil {
		http.Error(w, "Failed to create shift", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           id,
		"shift_number": shiftNumber,
		"created_at":   createdAt,
		"message":      "Shift created successfully",
	})
}

// ClosePOSShift closes a shift
func (h *ShiftHandler) ClosePOSShift(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	shiftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid shift ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ClosingAmount float64 `json:"closing_amount" validate:"required"`
		Notes         *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := h.db.Beginx()
	if err != nil {
		http.Error(w, "Failed to close shift", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Get shift details
	var shift RegisterShift
	err = tx.QueryRow("SELECT * FROM register_shifts WHERE id = $1 AND tenant_id = $2 AND status = 'open'",
		shiftID, tenantID).Scan(&shift.ID, &shift.TenantID, &shift.RegisterID, &shift.ShiftNumber,
		&shift.CashierID, &shift.OpeningBalance, &shift.ClosingBalance, &shift.ExpectedBalance,
		&shift.Variance, &shift.TotalSales, &shift.TotalCashSales, &shift.TotalCardSales,
		&shift.TotalReturns, &shift.TransactionCount, &shift.OpenedAt, &shift.ClosedAt,
		&shift.Status, &shift.Notes, &shift.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Active shift not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch shift", http.StatusInternalServerError)
		return
	}

	// Calculate expected balance
	expectedBalance := shift.OpeningBalance + shift.TotalSales
	variance := req.ClosingAmount - expectedBalance

	// Close shift
	_, err = tx.Exec(`
		UPDATE register_shifts 
		SET closing_balance = $1, expected_balance = $2, variance = $3, closed_at = $4, status = 'closed', notes = $5
		WHERE id = $6
	`, req.ClosingAmount, expectedBalance, variance, time.Now(), req.Notes, shiftID)

	if err != nil {
		http.Error(w, "Failed to close shift", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to close shift", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"closing_amount":   req.ClosingAmount,
		"expected_balance": expectedBalance,
		"variance":         variance,
		"message":          "Shift closed successfully",
	})
}
