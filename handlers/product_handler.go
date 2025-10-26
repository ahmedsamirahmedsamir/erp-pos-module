package pos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

// ProductHandler handles POS product-related operations
type ProductHandler struct {
	db          *sqlx.DB
	baseHandler *POSHandler
}

// NewProductHandler creates a new product handler
func NewProductHandler(db *sqlx.DB) *ProductHandler {
	return &ProductHandler{
		db:          db,
		baseHandler: NewPOSHandler(db),
	}
}

// GetPOSProducts retrieves POS-specific product settings
func (h *ProductHandler) GetPOSProducts(w http.ResponseWriter, r *http.Request) {
	_, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	registerID := r.URL.Query().Get("register_id")
	query := `
		SELECT pp.*, p.name as product_name, p.sku, p.selling_price
		FROM pos_products pp
		JOIN products p ON pp.product_id = p.id
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if registerID != "" {
		query += fmt.Sprintf(" AND pp.register_id = $%d", argIndex)
		args = append(args, registerID)
		argIndex++
	}

	query += " ORDER BY pp.display_order, pp.product_id"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch POS products", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []POSProduct
	for rows.Next() {
		var product POSProduct
		var productName, sku, quickKey, colorCode sql.NullString
		var sellingPrice sql.NullFloat64

		err := rows.Scan(
			&product.ID, &product.ProductID, &product.RegisterID, &product.IsAvailable,
			&product.DisplayOrder, &quickKey, &colorCode, &product.CreatedAt, &product.UpdatedAt,
			&productName, &sku, &sellingPrice,
		)
		if err != nil {
			continue
		}

		product.QuickKey = &quickKey.String
		product.ColorCode = &colorCode.String

		if productName.Valid {
			product.Product = &Product{
				ID:   product.ProductID,
				Name: productName.String,
				SKU:  sku.String,
			}
			if sellingPrice.Valid {
				product.Product.Price = sellingPrice.Float64
			}
		}

		products = append(products, product)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"products": products,
		"count":    len(products),
	})
}

// CreatePOSProduct links a product to POS
func (h *ProductHandler) CreatePOSProduct(w http.ResponseWriter, r *http.Request) {
	_, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		ProductID    int     `json:"product_id" validate:"required"`
		RegisterID   *int    `json:"register_id"`
		IsAvailable  bool    `json:"is_available"`
		DisplayOrder int     `json:"display_order"`
		QuickKey     *string `json:"quick_key"`
		ColorCode    *string `json:"color_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO pos_products (product_id, register_id, is_available, display_order, quick_key, color_code)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (product_id, register_id) DO UPDATE SET
			is_available = EXCLUDED.is_available,
			display_order = EXCLUDED.display_order,
			quick_key = EXCLUDED.quick_key,
			color_code = EXCLUDED.color_code,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	var id int
	var createdAt, updatedAt time.Time

	err = h.db.QueryRow(query, req.ProductID, req.RegisterID, req.IsAvailable, req.DisplayOrder, req.QuickKey, req.ColorCode).
		Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		http.Error(w, "Failed to create POS product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"created_at": createdAt,
		"updated_at": updatedAt,
		"message":    "POS product created successfully",
	})
}

// =================================================================
// QUICK SALE MANAGEMENT
// =================================================================

// GetQuickSaleCategories retrieves all quick sale categories
func (h *ProductHandler) GetQuickSaleCategories(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	query := `SELECT * FROM quick_sale_categories WHERE tenant_id = $1 AND is_active = true ORDER BY display_order, category_name`
	rows, err := h.db.Query(query, tenantID)
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []QuickSaleCategory
	for rows.Next() {
		var category QuickSaleCategory
		err := rows.Scan(&category.ID, &category.TenantID, &category.LocationID, &category.CategoryCode,
			&category.CategoryName, &category.ColorCode, &category.Icon, &category.DisplayOrder,
			&category.IsActive, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			continue
		}
		categories = append(categories, category)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categories,
		"count":      len(categories),
	})
}

// CreateQuickSaleCategory creates a new quick sale category
func (h *ProductHandler) CreateQuickSaleCategory(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	_ = tenantID // Suppress unused variable warning

	var req struct {
		CategoryCode string  `json:"category_code" validate:"required"`
		CategoryName string  `json:"category_name" validate:"required"`
		LocationID   *int    `json:"location_id"`
		ColorCode    *string `json:"color_code"`
		Icon         *string `json:"icon"`
		DisplayOrder int     `json:"display_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO quick_sale_categories (tenant_id, location_id, category_code, category_name, color_code, icon, display_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	var id int
	var createdAt time.Time

	err = h.db.QueryRow(query, tenantID, req.LocationID, req.CategoryCode, req.CategoryName,
		req.ColorCode, req.Icon, req.DisplayOrder).Scan(&id, &createdAt)
	if err != nil {
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"created_at": createdAt,
		"message":    "Quick sale category created successfully",
	})
}

// GetQuickSaleItems retrieves quick sale items
func (h *ProductHandler) GetQuickSaleItems(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	categoryID := r.URL.Query().Get("category_id")
	locationID := r.URL.Query().Get("location_id")

	query := `
		SELECT qsi.*, p.name as product_name, p.sku, p.selling_price
		FROM quick_sale_items qsi
		JOIN products p ON qsi.product_id = p.id
		WHERE qsi.tenant_id = $1 AND qsi.is_active = true
	`
	args := []interface{}{tenantID}
	argIndex := 2

	if categoryID != "" {
		query += fmt.Sprintf(" AND qsi.category_id = $%d", argIndex)
		args = append(args, categoryID)
		argIndex++
	}

	if locationID != "" {
		query += fmt.Sprintf(" AND qsi.location_id = $%d", argIndex)
		args = append(args, locationID)
		argIndex++
	}

	query += " ORDER BY qsi.display_order, qsi.id"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch quick sale items", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []QuickSaleItem
	for rows.Next() {
		var item QuickSaleItem
		var productName, sku sql.NullString
		var sellingPrice sql.NullFloat64

		err := rows.Scan(&item.ID, &item.TenantID, &item.LocationID, &item.CategoryID,
			&item.ProductID, &item.ButtonText, &item.ButtonColor, &item.DisplayOrder,
			&item.IsActive, &item.CreatedAt, &item.UpdatedAt, &productName, &sku, &sellingPrice)
		if err != nil {
			continue
		}

		if productName.Valid {
			item.Product = &Product{
				ID:   item.ProductID,
				Name: productName.String,
				SKU:  sku.String,
			}
			if sellingPrice.Valid {
				item.Product.Price = sellingPrice.Float64
			}
		}

		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}

// CreateQuickSaleItem creates a new quick sale item
func (h *ProductHandler) CreateQuickSaleItem(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.baseHandler.getTenantID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		ProductID    int     `json:"product_id" validate:"required"`
		CategoryID   *int    `json:"category_id"`
		LocationID   *int    `json:"location_id"`
		ButtonText   string  `json:"button_text" validate:"required"`
		ButtonColor  *string `json:"button_color"`
		DisplayOrder int     `json:"display_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO quick_sale_items (tenant_id, location_id, category_id, product_id, button_text, button_color, display_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, product_id, location_id) DO UPDATE SET
			button_text = EXCLUDED.button_text,
			button_color = EXCLUDED.button_color,
			display_order = EXCLUDED.display_order,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	var id int
	var createdAt, updatedAt time.Time

	err = h.db.QueryRow(query, tenantID, req.LocationID, req.CategoryID, req.ProductID,
		req.ButtonText, req.ButtonColor, req.DisplayOrder).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		http.Error(w, "Failed to create quick sale item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"created_at": createdAt,
		"updated_at": updatedAt,
		"message":    "Quick sale item created successfully",
	})
}
