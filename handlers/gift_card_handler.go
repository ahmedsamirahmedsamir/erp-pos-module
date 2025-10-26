package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// GiftCardHandler handles gift card operations
type GiftCardHandler struct {
	db          *sqlx.DB
	logger      *zap.Logger
	baseHandler *POSHandler
}

// NewGiftCardHandler creates a new gift card handler
func NewGiftCardHandler(db *sqlx.DB, logger *zap.Logger) *GiftCardHandler {
	return &GiftCardHandler{
		db:          db,
		logger:      logger,
		baseHandler: NewPOSHandler(db, logger),
	}
}

// GetGiftCards retrieves gift cards
func (h *GiftCardHandler) GetGiftCards(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	status := r.URL.Query().Get("status")
	customerID := r.URL.Query().Get("customer_id")

	query := `SELECT * FROM gift_cards WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argIndex := 2

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if customerID != "" {
		query += fmt.Sprintf(" AND purchased_by_customer_id = $%d", argIndex)
		args = append(args, customerID)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch gift cards", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cards []GiftCard
	for rows.Next() {
		var card GiftCard
		err := rows.Scan(&card.ID, &card.TenantID, &card.CompanyID, &card.CardNumber,
			&card.PINCode, &card.InitialValue, &card.CurrentBalance, &card.Currency,
			&card.PurchasedByCustomerID, &card.RecipientName, &card.RecipientEmail,
			&card.Status, &card.IssuedDate, &card.ExpiryDate, &card.LastUsedAt,
			&card.CreatedAt, &card.UpdatedAt)
		if err != nil {
			continue
		}
		cards = append(cards, card)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"gift_cards": cards,
		"count":      len(cards),
	})
}

// CreateGiftCard creates a new gift card
func (h *GiftCardHandler) CreateGiftCard(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		CompanyID             *string `json:"company_id"`
		InitialValue          float64 `json:"initial_value" validate:"required"`
		Currency              string  `json:"currency"`
		PurchasedByCustomerID *int    `json:"purchased_by_customer_id"`
		RecipientName         *string `json:"recipient_name"`
		RecipientEmail        *string `json:"recipient_email"`
		ExpiryDate            *string `json:"expiry_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}

	// Generate card number
	cardNumber := fmt.Sprintf("GC-%d", time.Now().UnixNano())

	// Parse expiry date if provided
	var expiryDate *time.Time
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		exp, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err == nil {
			expiryDate = &exp
		}
	}

	query := `
		INSERT INTO gift_cards (tenant_id, company_id, card_number, initial_value, current_balance,
		                       currency, purchased_by_customer_id, recipient_name, recipient_email,
		                       issued_date, expiry_date, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'active')
		RETURNING id, created_at, updated_at
	`

	var id int
	var createdAt, updatedAt time.Time

	err = h.db.QueryRow(query, tenantID, req.CompanyID, cardNumber, req.InitialValue, req.InitialValue,
		req.Currency, req.PurchasedByCustomerID, req.RecipientName, req.RecipientEmail,
		time.Now(), expiryDate).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		http.Error(w, "Failed to create gift card", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          id,
		"card_number": cardNumber,
		"created_at":  createdAt,
		"updated_at":  updatedAt,
		"message":     "Gift card created successfully",
	})
}

// RedeemGiftCard redeems a gift card for a transaction
func (h *GiftCardHandler) RedeemGiftCard(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		CardNumber    string  `json:"card_number" validate:"required"`
		Amount        float64 `json:"amount" validate:"required"`
		TransactionID *int    `json:"transaction_id"`
		Notes         *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, _ := h.baseHandler.getUserID(r)

	tx, err := h.db.Beginx()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Get gift card
	var giftCard GiftCard
	err = tx.QueryRow("SELECT * FROM gift_cards WHERE tenant_id = $1 AND card_number = $2 AND status = 'active'",
		tenantID, req.CardNumber).Scan(&giftCard.ID, &giftCard.TenantID, &giftCard.CompanyID, &giftCard.CardNumber,
		&giftCard.PINCode, &giftCard.InitialValue, &giftCard.CurrentBalance, &giftCard.Currency,
		&giftCard.PurchasedByCustomerID, &giftCard.RecipientName, &giftCard.RecipientEmail,
		&giftCard.Status, &giftCard.IssuedDate, &giftCard.ExpiryDate, &giftCard.LastUsedAt,
		&giftCard.CreatedAt, &giftCard.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Gift card not found or inactive", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch gift card", http.StatusInternalServerError)
		return
	}

	if giftCard.CurrentBalance < req.Amount {
		http.Error(w, "Insufficient gift card balance", http.StatusBadRequest)
		return
	}

	// Update gift card balance
	newBalance := giftCard.CurrentBalance - req.Amount
	_, err = tx.Exec("UPDATE gift_cards SET current_balance = $1, last_used_at = $2, updated_at = $3 WHERE id = $4",
		newBalance, time.Now(), time.Now(), giftCard.ID)
	if err != nil {
		http.Error(w, "Failed to update gift card", http.StatusInternalServerError)
		return
	}

	// Create gift card transaction
	_, err = tx.Exec(`
		INSERT INTO gift_card_transactions (tenant_id, gift_card_id, transaction_type, amount,
		                                   balance_before, balance_after, transaction_id, created_by, notes)
		VALUES ($1, $2, 'redeem', $3, $4, $5, $6, $7, $8)
	`, tenantID, giftCard.ID, req.Amount, giftCard.CurrentBalance, newBalance, req.TransactionID, userID, req.Notes)

	if err != nil {
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	// Mark card as used if balance is zero
	if newBalance <= 0 {
		_, err = tx.Exec("UPDATE gift_cards SET status = 'used' WHERE id = $1", giftCard.ID)
		if err != nil {
			http.Error(w, "Failed to update card status", http.StatusInternalServerError)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to redeem gift card", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"redeemed_amount":   req.Amount,
		"remaining_balance": newBalance,
		"message":           "Gift card redeemed successfully",
	})
}

// GetGiftCardByNumber retrieves a gift card by card number
func (h *GiftCardHandler) GetGiftCardByNumber(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	cardNumber := chi.URLParam(r, "number")

	var card GiftCard
	err = h.db.QueryRow("SELECT * FROM gift_cards WHERE tenant_id = $1 AND card_number = $2",
		tenantID, cardNumber).Scan(&card.ID, &card.TenantID, &card.CompanyID, &card.CardNumber,
		&card.PINCode, &card.InitialValue, &card.CurrentBalance, &card.Currency,
		&card.PurchasedByCustomerID, &card.RecipientName, &card.RecipientEmail,
		&card.Status, &card.IssuedDate, &card.ExpiryDate, &card.LastUsedAt,
		&card.CreatedAt, &card.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Gift card not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch gift card", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}
