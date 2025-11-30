package handlers

import (
	"auction/internal/middleware"
	"auction/internal/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type LotHandler struct {
	db *sql.DB
}

func NewLotHandler(db *sql.DB) *LotHandler {
	return &LotHandler{db: db}
}

func (h *LotHandler) GetLots(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT id, title, description, start_price, current_price, end_time, created_at
		FROM lots 
		WHERE status = 'active' 
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var lots []models.Lot
	for rows.Next() {
		var lot models.Lot
		err := rows.Scan(&lot.ID, &lot.Title, &lot.Description, &lot.StartPrice,
			&lot.CurrentPrice, &lot.EndTime, &lot.CreatedAt)
		if err != nil {
			http.Error(w, "lot search error", http.StatusInternalServerError)
			return
		}
		lots = append(lots, lot)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lots)
}

func (h *LotHandler) CreateLot(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	role, err := h.checkUserRole(user.ID)
	if role != "user" {
		http.Error(w, "no access", http.StatusUnauthorized)
		return
	}

	var lot models.Lot
	if err := json.NewDecoder(r.Body).Decode(&lot); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	//"2024-06-15T14:30:00+03:00"
	endTime, err := time.Parse(time.RFC3339, lot.EndTime)
	if err != nil {
		http.Error(w, "Invalid time format", http.StatusBadRequest)
		return
	}

	var lotID int
	createdAt := time.Now()
	err = h.db.QueryRow(
		`INSERT INTO lots (title, description, start_price, current_price, end_time, user_id, created_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		lot.Title, lot.Description, lot.StartPrice, lot.StartPrice, endTime, user.ID, createdAt,
	).Scan(&lotID)

	if err != nil {
		log.Printf("Error inserting new lot: %v", err)
		http.Error(w, "error creating lot", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "lot created successfully",
		"lot_id":  lotID,
	})
}

func (h *LotHandler) GetLot(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid lot ID", http.StatusBadRequest)
		return
	}

	var lot models.Lot

	err = h.db.QueryRow(`
		SELECT id, title, description, start_price, current_price, end_time, created_at 
		FROM lots WHERE id = $1
	`, id).Scan(&lot.ID, &lot.Title, &lot.Description, &lot.StartPrice,
		&lot.CurrentPrice, &lot.EndTime, &lot.CreatedAt)

	if err != nil {
		http.Error(w, "lot not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lot)
}

func (h *LotHandler) DeleteLot(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil || user.Role != "admin" {
		http.Error(w, "admin access required", http.StatusUnauthorized)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid lot ID", http.StatusBadRequest)
		return
	}
	_, err = h.db.Exec(`DELETE FROM lots WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "error deleting lot", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (h *LotHandler) checkUserRole(userID int) (string, error) {
	var role string
	err := h.db.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}
