package pos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// POSHandler handles all POS-related HTTP requests
type POSHandler struct {
	db *sqlx.DB
}

// NewPOSHandler creates a new POS handler
func NewPOSHandler(db *sqlx.DB) *POSHandler {
	return &POSHandler{db: db}
}

// =================================================================
// TENANT CONTEXT HELPERS
// =================================================================

// Helper functions to extract context from requests
// Note: These should be integrated with the ERP auth middleware
func (h *POSHandler) getTenantID(r *http.Request) (string, error) {
	// This should get the tenant from JWT or auth context
	// For now, using a placeholder
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		return "", fmt.Errorf("tenant context required")
	}
	return tenantID, nil
}

func (h *POSHandler) getUserID(r *http.Request) (int, error) {
	// This should get the user ID from JWT or auth context
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return 0, fmt.Errorf("user context required")
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID")
	}
	return userID, nil
}

// =================================================================
// SESSION MANAGEMENT
// =================================================================

// GetPOSSessions retrieves all POS sessions
func (h *POSHandler) GetPOSSessions(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	status := r.URL.Query().Get("status")
	registerID := r.URL.Query().Get("register_id")
	limit := r.URL.Query().Get("limit")

	if limit == "" {
		limit = "50"
	}

	query := `
		SELECT ps.*, pr.name as register_name, pr.code as register_code,
		       u.first_name as user_first_name, u.last_name as user_last_name
		FROM pos_sessions ps
		JOIN pos_registers pr ON ps.register_id = pr.id
		LEFT JOIN users u ON ps.user_id = u.id
		WHERE ps.tenant_id = $1
	`

	args := []interface{}{tenantID}
	argIndex := 2

	if status != "" {
		query += fmt.Sprintf(" AND ps.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if registerID != "" {
		query += fmt.Sprintf(" AND ps.register_id = $%d", argIndex)
		args = append(args, registerID)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY ps.session_start DESC LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch POS sessions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sessions []POSSession
	for rows.Next() {
		var session POSSession
		var registerName, registerCode sql.NullString
		var cashierFirstName, cashierLastName sql.NullString

		err := rows.Scan(
			&session.ID, &session.TenantID, &session.RegisterID, &session.CashDrawerID, &session.ShiftID,
			&session.UserID, &session.SessionNumber, &session.SessionStart, &session.SessionEnd,
			&session.OpeningAmount, &session.ClosingAmount, &session.TotalSales, &session.TotalRefunds,
			&session.TotalTransactions, &session.Status, &session.Notes, &session.Metadata,
			&session.CreatedAt, &session.UpdatedAt, &registerName, &registerCode,
			&cashierFirstName, &cashierLastName,
		)
		if err != nil {
			http.Error(w, "Failed to scan session", http.StatusInternalServerError)
			return
		}

		session.Register = &POSRegister{
			ID:   session.RegisterID,
			Name: registerName.String,
			Code: registerCode.String,
		}

		if cashierFirstName.Valid {
			var firstName, lastName string
			if cashierFirstName.Valid {
				firstName = cashierFirstName.String
			}
			if cashierLastName.Valid {
				lastName = cashierLastName.String
			}
			session.User = &User{
				ID:        session.UserID,
				FirstName: &firstName,
				LastName:  &lastName,
			}
		}

		sessions = append(sessions, session)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// CreatePOSSession creates a new POS session
func (h *POSHandler) CreatePOSSession(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		RegisterID    int     `json:"register_id" validate:"required"`
		OpeningAmount float64 `json:"opening_amount"`
		Notes         *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate session number
	sessionNumber := fmt.Sprintf("SES-%d", time.Now().Unix())

	// Check if there's an active session for this register
	var activeSessionID int
	err = h.db.QueryRow("SELECT id FROM pos_sessions WHERE tenant_id = $1 AND register_id = $2 AND status = 'active'",
		tenantID, req.RegisterID).Scan(&activeSessionID)
	if err == nil {
		http.Error(w, "There is already an active session for this register", http.StatusConflict)
		return
	}

	userID, _ := h.getUserID(r)

	// Start transaction
	tx, err := h.db.Beginx()
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Create POS session
	sessionQuery := `
		INSERT INTO pos_sessions (tenant_id, register_id, user_id, opening_amount, session_number, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, session_start, created_at, updated_at
	`

	var sessionID int
	var sessionStart, createdAt, updatedAt time.Time

	err = tx.QueryRow(sessionQuery, tenantID, req.RegisterID, userID, req.OpeningAmount, sessionNumber, req.Notes).
		Scan(&sessionID, &sessionStart, &createdAt, &updatedAt)

	if err != nil {
		http.Error(w, "Failed to create POS session", http.StatusInternalServerError)
		return
	}

	// Create cash drawer if opening amount > 0
	if req.OpeningAmount > 0 {
		drawerQuery := `
			INSERT INTO pos_cash_drawers (tenant_id, register_id, opening_amount, status, opened_by, opened_at)
			VALUES ($1, $2, $3, 'open', $4, $5)
			RETURNING id
		`

		var drawerID int
		err = tx.QueryRow(drawerQuery, tenantID, req.RegisterID, req.OpeningAmount, userID, sessionStart).Scan(&drawerID)
		if err != nil {
			http.Error(w, "Failed to create cash drawer", http.StatusInternalServerError)
			return
		}

		// Update session with cash drawer ID
		_, err = tx.Exec("UPDATE pos_sessions SET cash_drawer_id = $1 WHERE id = $2", drawerID, sessionID)
		if err != nil {
			http.Error(w, "Failed to update session with cash drawer", http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id":     sessionID,
		"session_number": sessionNumber,
		"session_start":  sessionStart,
		"created_at":     createdAt,
		"updated_at":     updatedAt,
		"message":        "POS session created successfully",
	})
}

// ClosePOSSession closes an active POS session
func (h *POSHandler) ClosePOSSession(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	sessionID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ClosingAmount float64 `json:"closing_amount"`
		Notes         *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, _ := h.getUserID(r)

	// Start transaction
	tx, err := h.db.Beginx()
	if err != nil {
		http.Error(w, "Failed to close session", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Get session details
	var session POSSession
	err = tx.QueryRow("SELECT * FROM pos_sessions WHERE id = $1 AND tenant_id = $2 AND status = 'active'",
		sessionID, tenantID).Scan(
		&session.ID, &session.TenantID, &session.RegisterID, &session.CashDrawerID, &session.ShiftID,
		&session.UserID, &session.SessionNumber, &session.SessionStart, &session.SessionEnd,
		&session.OpeningAmount, &session.ClosingAmount, &session.TotalSales, &session.TotalRefunds,
		&session.TotalTransactions, &session.Status, &session.Notes, &session.Metadata,
		&session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Active session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch session", http.StatusInternalServerError)
		return
	}

	// Close cash drawer if exists
	if session.CashDrawerID != nil {
		_, err = tx.Exec(`
			UPDATE pos_cash_drawers 
			SET closing_amount = $1, status = 'closed', closed_by = $2, closed_at = $3
			WHERE id = $4
		`, req.ClosingAmount, userID, time.Now(), *session.CashDrawerID)
		if err != nil {
			http.Error(w, "Failed to close cash drawer", http.StatusInternalServerError)
			return
		}
	}

	// Close session
	_, err = tx.Exec(`
		UPDATE pos_sessions 
		SET closing_amount = $1, session_end = $2, status = 'closed', notes = $3
		WHERE id = $4
	`, req.ClosingAmount, time.Now(), req.Notes, sessionID)
	if err != nil {
		http.Error(w, "Failed to close session", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to close session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "POS session closed successfully",
	})
}

// =================================================================
// TRANSACTION MANAGEMENT
// =================================================================

// GetPOSTransactions retrieves POS transactions
func (h *POSHandler) GetPOSTransactions(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	registerID := r.URL.Query().Get("register_id")
	status := r.URL.Query().Get("status")
	limit := r.URL.Query().Get("limit")

	if limit == "" {
		limit = "100"
	}

	query := `
		SELECT pt.*, c.first_name, c.last_name, c.company_name,
		       u.first_name as cashier_first_name, u.last_name as cashier_last_name
		FROM pos_transactions pt
		LEFT JOIN customers c ON pt.customer_id = c.id
		LEFT JOIN users u ON pt.cashier_id = u.id
		WHERE pt.tenant_id = $1
	`

	args := []interface{}{tenantID}
	argIndex := 2

	if sessionID != "" {
		query += fmt.Sprintf(" AND pt.session_id = $%d", argIndex)
		args = append(args, sessionID)
		argIndex++
	}

	if registerID != "" {
		query += fmt.Sprintf(" AND pt.register_id = $%d", argIndex)
		args = append(args, registerID)
		argIndex++
	}

	if status != "" {
		query += fmt.Sprintf(" AND pt.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY pt.transaction_date DESC LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch POS transactions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []POSTransaction
	for rows.Next() {
		var transaction POSTransaction
		var firstName, lastName, companyName sql.NullString
		var cashierFirstName, cashierLastName sql.NullString
		var customFieldsJSON sql.NullString

		err := rows.Scan(
			&transaction.ID, &transaction.TenantID, &transaction.TransactionNumber,
			&transaction.SessionID, &transaction.RegisterID, &transaction.ShiftID,
			&transaction.CustomerID, &transaction.TransactionDate, &transaction.TransactionType,
			&transaction.Status, &transaction.Subtotal, &transaction.TaxAmount,
			&transaction.DiscountAmount, &transaction.TipAmount, &transaction.TotalAmount,
			&transaction.ChangeAmount, &transaction.CashierID, &transaction.ManagerID,
			&transaction.Notes, &customFieldsJSON, &transaction.CreatedAt, &transaction.UpdatedAt,
			&firstName, &lastName, &companyName, &cashierFirstName, &cashierLastName,
		)
		if err != nil {
			continue
		}

		if firstName.Valid && transaction.CustomerID != nil {
			transaction.Customer = &Customer{
				ID:        *transaction.CustomerID,
				FirstName: &firstName.String,
				LastName:  &lastName.String,
			}
		}

		if cashierFirstName.Valid {
			var fname, lname string
			if cashierFirstName.Valid {
				fname = cashierFirstName.String
			}
			if cashierLastName.Valid {
				lname = cashierLastName.String
			}
			transaction.Cashier = &User{
				ID:        transaction.CashierID,
				FirstName: &fname,
				LastName:  &lname,
			}
		}

		if customFieldsJSON.Valid {
			json.Unmarshal([]byte(customFieldsJSON.String), &transaction.CustomFields)
		}

		transactions = append(transactions, transaction)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": transactions,
		"count":        len(transactions),
	})
}

// CreatePOSTransaction creates a new POS transaction
func (h *POSHandler) CreatePOSTransaction(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		SessionID      int                    `json:"session_id" validate:"required"`
		RegisterID     int                    `json:"register_id" validate:"required"`
		CustomerID     *int                   `json:"customer_id"`
		Subtotal       float64                `json:"subtotal" validate:"required"`
		TaxAmount      float64                `json:"tax_amount"`
		DiscountAmount float64                `json:"discount_amount"`
		TipAmount      float64                `json:"tip_amount"`
		TotalAmount    float64                `json:"total_amount" validate:"required"`
		ChangeAmount   float64                `json:"change_amount"`
		Notes          *string                `json:"notes"`
		Items          []POSTransactionItem   `json:"items" validate:"required"`
		Payments       []POSPayment           `json:"payments" validate:"required"`
		CustomFields   map[string]interface{} `json:"custom_fields"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "At least one item is required", http.StatusBadRequest)
		return
	}

	if len(req.Payments) == 0 {
		http.Error(w, "At least one payment is required", http.StatusBadRequest)
		return
	}

	// Generate transaction number
	transactionNumber := fmt.Sprintf("TXN-%d", time.Now().Unix())

	userID, _ := h.getUserID(r)

	// Start transaction
	tx, err := h.db.Beginx()
	if err != nil {
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Create POS transaction
	transactionQuery := `
		INSERT INTO pos_transactions (tenant_id, transaction_number, session_id, register_id, customer_id,
		                             subtotal, tax_amount, discount_amount, tip_amount, total_amount,
		                             change_amount, cashier_id, notes, custom_fields)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, transaction_date, created_at, updated_at
	`

	var transactionID int
	var transactionDate, createdAt, updatedAt time.Time

	customFieldsJSON, _ := json.Marshal(req.CustomFields)

	err = tx.QueryRow(transactionQuery, tenantID, transactionNumber, req.SessionID, req.RegisterID, req.CustomerID,
		req.Subtotal, req.TaxAmount, req.DiscountAmount, req.TipAmount, req.TotalAmount,
		req.ChangeAmount, userID, req.Notes, customFieldsJSON).Scan(&transactionID, &transactionDate, &createdAt, &updatedAt)

	if err != nil {
		http.Error(w, "Failed to create POS transaction", http.StatusInternalServerError)
		return
	}

	// Create transaction items
	for _, item := range req.Items {
		itemQuery := `
			INSERT INTO pos_transaction_items (transaction_id, product_id, quantity, unit_price,
			                                   discount_percent, discount_amount, tax_rate, tax_amount, notes, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`

		itemMetadata, _ := json.Marshal(map[string]interface{}{})
		_, err = tx.Exec(itemQuery, transactionID, item.ProductID, item.Quantity, item.UnitPrice,
			item.DiscountPercent, item.DiscountAmount, item.TaxRate, item.TaxAmount, item.Notes, itemMetadata)
		if err != nil {
			http.Error(w, "Failed to create transaction item", http.StatusInternalServerError)
			return
		}
	}

	// Create payments
	for _, payment := range req.Payments {
		paymentQuery := `
			INSERT INTO pos_payments (transaction_id, payment_method, amount, reference_number, status, card_type, notes, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`

		paymentMetadata, _ := json.Marshal(map[string]interface{}{})
		_, err = tx.Exec(paymentQuery, transactionID, payment.PaymentMethod, payment.Amount,
			payment.ReferenceNumber, "completed", payment.CardType, payment.Notes, paymentMetadata)
		if err != nil {
			http.Error(w, "Failed to create payment", http.StatusInternalServerError)
			return
		}
	}

	// Update session totals
	_, err = tx.Exec(`
		UPDATE pos_sessions 
		SET total_sales = total_sales + $1, total_transactions = total_transactions + 1
		WHERE id = $2
	`, req.TotalAmount, req.SessionID)
	if err != nil {
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transaction_id":     transactionID,
		"transaction_number": transactionNumber,
		"transaction_date":   transactionDate,
		"created_at":         createdAt,
		"updated_at":         updatedAt,
		"message":            "POS transaction created successfully",
	})
}

// GetPOSTransaction retrieves a single POS transaction by ID
func (h *POSHandler) GetPOSTransaction(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	// Get transaction details
	query := `
		SELECT pt.*, c.first_name, c.last_name, c.company_name,
		       u.first_name as cashier_first_name, u.last_name as cashier_last_name
		FROM pos_transactions pt
		LEFT JOIN customers c ON pt.customer_id = c.id
		LEFT JOIN users u ON pt.cashier_id = u.id
		WHERE pt.id = $1 AND pt.tenant_id = $2
	`

	var transaction POSTransaction
	var firstName, lastName, companyName sql.NullString
	var cashierFirstName, cashierLastName sql.NullString
	var customFieldsJSON sql.NullString

	err = h.db.QueryRow(query, id, tenantID).Scan(
		&transaction.ID, &transaction.TenantID, &transaction.TransactionNumber,
		&transaction.SessionID, &transaction.RegisterID, &transaction.ShiftID,
		&transaction.CustomerID, &transaction.TransactionDate, &transaction.TransactionType,
		&transaction.Status, &transaction.Subtotal, &transaction.TaxAmount,
		&transaction.DiscountAmount, &transaction.TipAmount, &transaction.TotalAmount,
		&transaction.ChangeAmount, &transaction.CashierID, &transaction.ManagerID,
		&transaction.Notes, &customFieldsJSON, &transaction.CreatedAt, &transaction.UpdatedAt,
		&firstName, &lastName, &companyName, &cashierFirstName, &cashierLastName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Transaction not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch transaction", http.StatusInternalServerError)
		return
	}

	// Build customer info
	if firstName.Valid && transaction.CustomerID != nil {
		transaction.Customer = &Customer{
			ID:        *transaction.CustomerID,
			FirstName: &firstName.String,
			LastName:  &lastName.String,
		}
	}

	if cashierFirstName.Valid {
		var fname, lname string
		if cashierFirstName.Valid {
			fname = cashierFirstName.String
		}
		if cashierLastName.Valid {
			lname = cashierLastName.String
		}
		transaction.Cashier = &User{
			ID:        transaction.CashierID,
			FirstName: &fname,
			LastName:  &lname,
		}
	}

	if customFieldsJSON.Valid {
		json.Unmarshal([]byte(customFieldsJSON.String), &transaction.CustomFields)
	}

	// Get transaction items
	itemsQuery := `
		SELECT pti.*, p.name as product_name, p.sku
		FROM pos_transaction_items pti
		JOIN products p ON pti.product_id = p.id
		WHERE pti.transaction_id = $1
		ORDER BY pti.id
	`

	itemRows, err := h.db.Query(itemsQuery, id)
	if err == nil {
		defer itemRows.Close()

		for itemRows.Next() {
			var item POSTransactionItem
			var productName, sku sql.NullString
			var metadataJSON sql.NullString

			err := itemRows.Scan(
				&item.ID, &item.TransactionID, &item.ProductID, &item.Quantity,
				&item.UnitPrice, &item.DiscountPercent, &item.DiscountAmount,
				&item.TaxRate, &item.TaxAmount, &item.LineTotal, &item.Notes,
				&metadataJSON, &item.CreatedAt, &productName, &sku,
			)
			if err != nil {
				continue
			}

			item.Product = &Product{
				ID:   item.ProductID,
				Name: productName.String,
				SKU:  sku.String,
			}

			if metadataJSON.Valid {
				json.Unmarshal([]byte(metadataJSON.String), &item.Metadata)
			}

			transaction.Items = append(transaction.Items, item)
		}
	}

	// Get payments
	paymentsQuery := `
		SELECT id, transaction_id, payment_method, amount, reference_number, card_type, status, processed_at, notes, metadata, created_at
		FROM pos_payments
		WHERE transaction_id = $1
		ORDER BY id
	`

	paymentRows, err := h.db.Query(paymentsQuery, id)
	if err == nil {
		defer paymentRows.Close()

		for paymentRows.Next() {
			var payment POSPayment
			var metadataJSON sql.NullString

			err := paymentRows.Scan(
				&payment.ID, &payment.TransactionID, &payment.PaymentMethod,
				&payment.Amount, &payment.ReferenceNumber, &payment.CardType, &payment.Status,
				&payment.ProcessedAt, &payment.Notes, &metadataJSON, &payment.CreatedAt,
			)
			if err != nil {
				continue
			}

			if metadataJSON.Valid {
				json.Unmarshal([]byte(metadataJSON.String), &payment.Metadata)
			}

			transaction.Payments = append(transaction.Payments, payment)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

// =================================================================
// RECEIPT MANAGEMENT
// =================================================================

// CreateReceipt creates a receipt for a transaction
func (h *POSHandler) CreateReceipt(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		TransactionID int    `json:"transaction_id" validate:"required"`
		ReceiptType   string `json:"receipt_type" validate:"required"`
		ReceiptData   string `json:"receipt_data" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate receipt number
	receiptNumber := fmt.Sprintf("RCP-%d", time.Now().Unix())

	query := `
		INSERT INTO pos_receipts (tenant_id, transaction_id, receipt_number, receipt_type, receipt_data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	var receiptID int
	var createdAt time.Time

	err = h.db.QueryRow(query, tenantID, req.TransactionID, receiptNumber, req.ReceiptType, req.ReceiptData).
		Scan(&receiptID, &createdAt)

	if err != nil {
		http.Error(w, "Failed to create receipt", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"receipt_id":     receiptID,
		"receipt_number": receiptNumber,
		"created_at":     createdAt,
		"message":        "Receipt created successfully",
	})
}

// PrintReceipt marks a receipt as printed
func (h *POSHandler) PrintReceipt(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	receiptID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid receipt ID", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE pos_receipts 
		SET printed_at = $1, reprint_count = reprint_count + 1
		WHERE id = $2
		RETURNING receipt_number, printed_at, reprint_count
	`

	var receiptNumber string
	var printedAt time.Time
	var reprintCount int

	err = h.db.QueryRow(query, time.Now(), receiptID).Scan(&receiptNumber, &printedAt, &reprintCount)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Receipt not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to print receipt", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"receipt_number": receiptNumber,
		"printed_at":     printedAt,
		"reprint_count":  reprintCount,
		"message":        "Receipt printed successfully",
	})
}

// =================================================================
// REGISTER MANAGEMENT
// =================================================================

// GetPOSRegisters retrieves all POS registers
func (h *POSHandler) GetPOSRegisters(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	query := `
		SELECT id, tenant_id, company_id, name, code, location_id, register_type, status,
		       opening_balance, current_balance, expected_balance, opened_at, opened_by,
		       closed_at, closed_by, is_active, metadata, created_at, updated_at
		FROM pos_registers
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY name
	`

	rows, err := h.db.Query(query, tenantID)
	if err != nil {
		http.Error(w, "Failed to fetch POS registers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var registers []POSRegister
	for rows.Next() {
		var register POSRegister
		var metadataJSON sql.NullString

		err := rows.Scan(
			&register.ID, &register.TenantID, &register.CompanyID, &register.Name, &register.Code,
			&register.LocationID, &register.RegisterType, &register.Status,
			&register.OpeningBalance, &register.CurrentBalance, &register.ExpectedBalance,
			&register.OpenedAt, &register.OpenedBy, &register.ClosedAt, &register.ClosedBy,
			&register.IsActive, &metadataJSON, &register.CreatedAt, &register.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &register.Metadata)
		}

		registers = append(registers, register)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"registers": registers,
		"count":     len(registers),
	})
}

// CreatePOSRegister creates a new POS register
func (h *POSHandler) CreatePOSRegister(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		Name         string  `json:"name" validate:"required"`
		Code         string  `json:"code" validate:"required"`
		LocationID   *int    `json:"location_id"`
		RegisterType string  `json:"register_type"`
		CompanyID    *string `json:"company_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Code == "" {
		http.Error(w, "Name and code are required", http.StatusBadRequest)
		return
	}

	if req.RegisterType == "" {
		req.RegisterType = "main"
	}

	query := `
		INSERT INTO pos_registers (tenant_id, name, code, location_id, register_type, company_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	var id int
	var createdAt, updatedAt time.Time

	err = h.db.QueryRow(query, tenantID, req.Name, req.Code, req.LocationID, req.RegisterType, req.CompanyID).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		http.Error(w, "Failed to create POS register", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"created_at": createdAt,
		"updated_at": updatedAt,
		"message":    "POS register created successfully",
	})
}

// =================================================================
// ANALYTICS & DASHBOARD
// =================================================================

// GetPOSAnalytics retrieves comprehensive POS analytics
func (h *POSHandler) GetPOSAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	registerID := r.URL.Query().Get("register_id")

	if startDate == "" || endDate == "" {
		http.Error(w, "Start date and end date are required", http.StatusBadRequest)
		return
	}

	// Get daily sales summary
	salesQuery := `
		SELECT 
			DATE(pt.created_at) as sale_date,
			COUNT(DISTINCT pt.id) as transaction_count,
			SUM(pt.total_amount) as total_sales,
			AVG(pt.total_amount) as average_transaction_value,
			SUM(pt.tax_amount) as total_tax,
			SUM(pt.discount_amount) as total_discounts
		FROM pos_transactions pt
		WHERE pt.tenant_id = $1 AND pt.created_at BETWEEN $2 AND $3
		  AND pt.status = 'completed'
	`

	args := []interface{}{tenantID, startDate, endDate}
	argIndex := 4

	if registerID != "" {
		salesQuery += fmt.Sprintf(" AND pt.register_id = $%d", argIndex)
		args = append(args, registerID)
		argIndex++
	}

	salesQuery += " GROUP BY DATE(pt.created_at) ORDER BY sale_date"

	rows, err := h.db.Query(salesQuery, args...)
	if err != nil {
		http.Error(w, "Failed to fetch POS analytics", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var analytics struct {
		Period                  string    `json:"period"`
		GeneratedAt             time.Time `json:"generated_at"`
		TotalTransactions       int       `json:"total_transactions"`
		TotalSales              float64   `json:"total_sales"`
		TotalTax                float64   `json:"total_tax"`
		TotalDiscounts          float64   `json:"total_discounts"`
		AverageTransactionValue float64   `json:"average_transaction_value"`
	}

	for rows.Next() {
		var saleDate time.Time
		var transactionCount int
		var totalSales, avgValue, totalTax, totalDiscounts float64

		err := rows.Scan(&saleDate, &transactionCount, &totalSales, &avgValue, &totalTax, &totalDiscounts)
		if err != nil {
			continue
		}

		analytics.TotalTransactions += transactionCount
		analytics.TotalSales += totalSales
		analytics.TotalTax += totalTax
		analytics.TotalDiscounts += totalDiscounts
	}

	if analytics.TotalTransactions > 0 {
		analytics.AverageTransactionValue = analytics.TotalSales / float64(analytics.TotalTransactions)
	}

	analytics.Period = fmt.Sprintf("%s to %s", startDate, endDate)
	analytics.GeneratedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}
