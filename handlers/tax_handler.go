package main

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// TaxHandler handles tax-related operations
type TaxHandler struct {
	db          *sqlx.DB
	logger      *zap.Logger
	baseHandler *POSHandler
}

// NewTaxHandler creates a new tax handler
func NewTaxHandler(db *sqlx.DB, logger *zap.Logger) *TaxHandler {
	return &TaxHandler{
		db:          db,
		logger:      logger,
		baseHandler: NewPOSHandler(db, logger),
	}
}

// GetPOSTaxes retrieves all tax rates
func (h *TaxHandler) GetPOSTaxes(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	query := `SELECT * FROM pos_taxes WHERE tenant_id = $1 ORDER BY tax_type, name`
	rows, err := h.db.Query(query, tenantID)
	if err != nil {
		http.Error(w, "Failed to fetch taxes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var taxes []POSTax
	for rows.Next() {
		var tax POSTax
		err := rows.Scan(&tax.ID, &tax.TenantID, &tax.Name, &tax.Code, &tax.Rate,
			&tax.TaxType, &tax.Description, &tax.IsActive, &tax.CreatedAt, &tax.UpdatedAt)
		if err != nil {
			continue
		}
		taxes = append(taxes, tax)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"taxes": taxes,
		"count": len(taxes),
	})
}

// CreatePOSTax creates a new tax rate
func (h *TaxHandler) CreatePOSTax(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		Name        string  `json:"name" validate:"required"`
		Code        string  `json:"code" validate:"required"`
		Rate        float64 `json:"rate" validate:"required"`
		TaxType     string  `json:"tax_type"`
		Description *string `json:"description"`
		IsActive    bool    `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TaxType == "" {
		req.TaxType = "sales"
	}

	query := `
		INSERT INTO pos_taxes (tenant_id, name, code, rate, tax_type, description, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	var id int
	// var createdAt, updatedAt time.Time

	err = h.db.QueryRow(query, tenantID, req.Name, req.Code, req.Rate, req.TaxType, req.Description, req.IsActive).
		Scan(&id)
	if err != nil {
		http.Error(w, "Failed to create tax", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "Tax created successfully",
	})
}
