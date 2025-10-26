package pos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

// DiscountHandler handles discount and coupon operations
type DiscountHandler struct {
	db          *sqlx.DB
	baseHandler *POSHandler
}

// NewDiscountHandler creates a new discount handler
func NewDiscountHandler(db *sqlx.DB) *DiscountHandler {
	return &DiscountHandler{
		db:          db,
		baseHandler: NewPOSHandler(db),
	}
}

// GetDiscountRules retrieves all discount rules
func (h *DiscountHandler) GetDiscountRules(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	isActive := r.URL.Query().Get("is_active")

	query := `SELECT * FROM discount_rules WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argIndex := 2

	if isActive == "true" {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, true)
		argIndex++
	}

	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch discount rules", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rules []DiscountRule
	for rows.Next() {
		var rule DiscountRule
		err := rows.Scan(&rule.ID, &rule.TenantID, &rule.CompanyID, &rule.RuleCode, &rule.RuleName,
			&rule.DiscountType, &rule.DiscountValue, &rule.AppliesTo, &rule.MinPurchaseAmount,
			&rule.MaxDiscountAmount, &rule.BuyQuantity, &rule.GetQuantity, &rule.CustomerGroupID,
			&rule.ValidFrom, &rule.ValidTo, &rule.DaysOfWeek, &rule.TimeFrom, &rule.TimeTo,
			&rule.UsageLimit, &rule.UsageCount, &rule.RequiresApproval, &rule.IsActive,
			&rule.Priority, &rule.CreatedAt, &rule.UpdatedAt)
		if err != nil {
			continue
		}
		rules = append(rules, rule)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

// CreateDiscountRule creates a new discount rule
func (h *DiscountHandler) CreateDiscountRule(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		CompanyID         *string  `json:"company_id"`
		RuleCode          string   `json:"rule_code" validate:"required"`
		RuleName          string   `json:"rule_name" validate:"required"`
		DiscountType      string   `json:"discount_type" validate:"required"`
		DiscountValue     float64  `json:"discount_value" validate:"required"`
		AppliesTo         string   `json:"applies_to" validate:"required"`
		MinPurchaseAmount *float64 `json:"min_purchase_amount"`
		MaxDiscountAmount *float64 `json:"max_discount_amount"`
		BuyQuantity       *int     `json:"buy_quantity"`
		GetQuantity       *int     `json:"get_quantity"`
		CustomerGroupID   *int     `json:"customer_group_id"`
		ValidFrom         string   `json:"valid_from" validate:"required"`
		ValidTo           *string  `json:"valid_to"`
		DaysOfWeek        []int    `json:"days_of_week"`
		TimeFrom          *string  `json:"time_from"`
		TimeTo            *string  `json:"time_to"`
		UsageLimit        *int     `json:"usage_limit"`
		RequiresApproval  bool     `json:"requires_approval"`
		IsActive          bool     `json:"is_active"`
		Priority          int      `json:"priority"`
		ProductIDs        []int    `json:"product_ids"`
		CategoryIDs       []int    `json:"category_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validFrom, err := time.Parse("2006-01-02 15:04:05", req.ValidFrom)
	if err != nil {
		validFrom, _ = time.Parse("2006-01-02", req.ValidFrom)
	}

	var validTo *time.Time
	if req.ValidTo != nil && *req.ValidTo != "" {
		vt, err := time.Parse("2006-01-02 15:04:05", *req.ValidTo)
		if err != nil {
			vt, _ = time.Parse("2006-01-02", *req.ValidTo)
		}
		validTo = &vt
	}

	tx, err := h.db.Beginx()
	if err != nil {
		http.Error(w, "Failed to create discount rule", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert discount rule
	ruleQuery := `
		INSERT INTO discount_rules (tenant_id, company_id, rule_code, rule_name, discount_type, discount_value,
		                           applies_to, min_purchase_amount, max_discount_amount, buy_quantity, get_quantity,
		                           customer_group_id, valid_from, valid_to, days_of_week, time_from, time_to,
		                           usage_limit, requires_approval, is_active, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id, created_at, updated_at
	`

	var ruleID int
	var createdAt, updatedAt time.Time

	err = tx.QueryRow(ruleQuery, tenantID, req.CompanyID, req.RuleCode, req.RuleName, req.DiscountType,
		req.DiscountValue, req.AppliesTo, req.MinPurchaseAmount, req.MaxDiscountAmount, req.BuyQuantity,
		req.GetQuantity, req.CustomerGroupID, validFrom, validTo, req.DaysOfWeek, req.TimeFrom,
		req.TimeTo, req.UsageLimit, req.RequiresApproval, req.IsActive, req.Priority).
		Scan(&ruleID, &createdAt, &updatedAt)

	if err != nil {
		http.Error(w, "Failed to create discount rule", http.StatusInternalServerError)
		return
	}

	// Add products if specified
	for _, productID := range req.ProductIDs {
		_, err = tx.Exec("INSERT INTO discount_rule_products (discount_rule_id, product_id, category_id) VALUES ($1, $2, NULL)",
			ruleID, productID)
		if err != nil {
			http.Error(w, "Failed to add product to rule", http.StatusInternalServerError)
			return
		}
	}

	// Add categories if specified
	for _, categoryID := range req.CategoryIDs {
		_, err = tx.Exec("INSERT INTO discount_rule_products (discount_rule_id, product_id, category_id) VALUES ($1, NULL, $2)",
			ruleID, categoryID)
		if err != nil {
			http.Error(w, "Failed to add category to rule", http.StatusInternalServerError)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to create discount rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         ruleID,
		"created_at": createdAt,
		"updated_at": updatedAt,
		"message":    "Discount rule created successfully",
	})
}

// ValidateCoupon validates a coupon code
func (h *DiscountHandler) ValidateCoupon(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		CouponCode string  `json:"coupon_code" validate:"required"`
		Amount     float64 `json:"amount" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if coupon exists and is valid
	var coupon CouponCode
	query := `SELECT * FROM coupon_codes WHERE tenant_id = $1 AND coupon_code = $2 AND is_active = true`
	err = h.db.QueryRow(query, tenantID, req.CouponCode).Scan(&coupon.ID, &coupon.TenantID, &coupon.DiscountRuleID,
		&coupon.CouponCode, &coupon.Description, &coupon.MaxUses, &coupon.MaxUsesPerCustomer,
		&coupon.CurrentUses, &coupon.ValidFrom, &coupon.ValidTo, &coupon.IsActive, &coupon.CreatedAt, &coupon.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid or inactive coupon", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to validate coupon", http.StatusInternalServerError)
		return
	}

	// Check validity
	now := time.Now()
	if now.Before(coupon.ValidFrom) {
		http.Error(w, "Coupon is not yet valid", http.StatusBadRequest)
		return
	}

	if coupon.ValidTo != nil && now.After(*coupon.ValidTo) {
		http.Error(w, "Coupon has expired", http.StatusBadRequest)
		return
	}

	if coupon.MaxUses != nil && coupon.CurrentUses >= *coupon.MaxUses {
		http.Error(w, "Coupon usage limit exceeded", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":            true,
		"coupon_id":        coupon.ID,
		"discount_rule_id": coupon.DiscountRuleID,
		"description":      coupon.Description,
		"message":          "Coupon is valid",
	})
}

// CreateCouponCode creates a new coupon
func (h *DiscountHandler) CreateCouponCode(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		DiscountRuleID     *int    `json:"discount_rule_id"`
		CouponCode         string  `json:"coupon_code" validate:"required"`
		Description        *string `json:"description"`
		MaxUses            *int    `json:"max_uses"`
		MaxUsesPerCustomer *int    `json:"max_uses_per_customer"`
		ValidFrom          string  `json:"valid_from" validate:"required"`
		ValidTo            *string `json:"valid_to"`
		IsActive           bool    `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validFrom, err := time.Parse("2006-01-02 15:04:05", req.ValidFrom)
	if err != nil {
		validFrom, _ = time.Parse("2006-01-02", req.ValidFrom)
	}

	var validTo *time.Time
	if req.ValidTo != nil && *req.ValidTo != "" {
		vt, err := time.Parse("2006-01-02 15:04:05", *req.ValidTo)
		if err != nil {
			vt, _ = time.Parse("2006-01-02", *req.ValidTo)
		}
		validTo = &vt
	}

	query := `
		INSERT INTO coupon_codes (tenant_id, discount_rule_id, coupon_code, description,
		                         max_uses, max_uses_per_customer, valid_from, valid_to, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	var id int
	var createdAt, updatedAt time.Time

	err = h.db.QueryRow(query, tenantID, req.DiscountRuleID, req.CouponCode, req.Description,
		req.MaxUses, req.MaxUsesPerCustomer, validFrom, validTo, req.IsActive).
		Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		http.Error(w, "Failed to create coupon code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"created_at": createdAt,
		"updated_at": updatedAt,
		"message":    "Coupon code created successfully",
	})
}


