package pos

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

// CustomerHandler handles POS customer operations
type CustomerHandler struct {
	db          *sqlx.DB
	baseHandler *POSHandler
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(db *sqlx.DB) *CustomerHandler {
	return &CustomerHandler{
		db:          db,
		baseHandler: NewPOSHandler(db),
	}
}

// GetPOSCustomers retrieves POS-specific customer information
func (h *CustomerHandler) GetPOSCustomers(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT pc.*, c.first_name, c.last_name, c.email, c.phone
		FROM pos_customers pc
		JOIN customers c ON pc.customer_id = c.id
		ORDER BY pc.total_spent DESC
	`
	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, "Failed to fetch POS customers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type CustomerRow struct {
		ID            int
		CustomerID    int
		LoyaltyPoints int
		TotalSpent    float64
		LastVisit     sql.NullTime
		IsVIP         bool
		CreatedAt     time.Time
		UpdatedAt     time.Time
		FirstName     sql.NullString
		LastName      sql.NullString
		Email         string
		Phone         string
	}

	var customers []CustomerRow

	for rows.Next() {
		var customer CustomerRow

		err := rows.Scan(&customer.ID, &customer.CustomerID, &customer.LoyaltyPoints,
			&customer.TotalSpent, &customer.LastVisit, &customer.IsVIP, &customer.CreatedAt,
			&customer.UpdatedAt, &customer.FirstName, &customer.LastName, &customer.Email,
			&customer.Phone)
		if err != nil {
			continue
		}

		customers = append(customers, customer)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"customers": customers,
		"count":     len(customers),
	})
}

// GetCustomerLoyalty retrieves loyalty points for a customer
func (h *CustomerHandler) GetCustomerLoyalty(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	if customerID == "" {
		http.Error(w, "customer_id required", http.StatusBadRequest)
		return
	}

	var customer struct {
		POSCustomer
		Customer *Customer
	}

	err := h.db.QueryRow(`
		SELECT pc.*, c.first_name, c.last_name, c.email, c.phone
		FROM pos_customers pc
		JOIN customers c ON pc.customer_id = c.id
		WHERE pc.customer_id = $1
	`, customerID).Scan(&customer.ID, &customer.CustomerID, &customer.LoyaltyPoints,
		&customer.TotalSpent, &customer.LastVisit, &customer.IsVIP, &customer.CreatedAt,
		&customer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch loyalty info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

// UpdateLoyaltyPoints updates customer loyalty points
func (h *CustomerHandler) UpdateLoyaltyPoints(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID  int     `json:"customer_id" validate:"required"`
		PointsToAdd int     `json:"points_to_add"`
		Notes       *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE pos_customers 
		SET loyalty_points = loyalty_points + $1, updated_at = NOW()
		WHERE customer_id = $2
		RETURNING loyalty_points
	`

	var newBalance int
	err := h.db.QueryRow(query, req.PointsToAdd, req.CustomerID).Scan(&newBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create entry if doesn't exist
			_, err = h.db.Exec(`
				INSERT INTO pos_customers (customer_id, loyalty_points, total_spent, last_visit)
				VALUES ($1, $2, 0, NOW())
			`, req.CustomerID, req.PointsToAdd)
			if err != nil {
				http.Error(w, "Failed to update loyalty points", http.StatusInternalServerError)
				return
			}
			newBalance = req.PointsToAdd
		} else {
			http.Error(w, "Failed to update loyalty points", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"customer_id":  req.CustomerID,
		"points_added": req.PointsToAdd,
		"new_balance":  newBalance,
		"message":      "Loyalty points updated successfully",
	})
}
