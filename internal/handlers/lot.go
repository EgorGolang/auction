package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type LotHandler struct {
	DB *sql.DB
}

func NewLotHandler(db *sql.DB) *LotHandler {
	return &LotHandler{DB: db}
}

// GetLots - получение списка лотов
func (h *LotHandler) GetLots(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, title, description, start_price, current_price, end_time, status 
		FROM lots 
		WHERE status = 'active' 
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var lots []map[string]interface{}
	for rows.Next() {
		var lot struct {
			ID           int       `json:"id"`
			Title        string    `json:"title"`
			Description  string    `json:"description"`
			StartPrice   float64   `json:"start_price"`
			CurrentPrice float64   `json:"current_price"`
			EndTime      time.Time `json:"end_time"`
			Status       string    `json:"status"`
		}

		err := rows.Scan(&lot.ID, &lot.Title, &lot.Description, &lot.StartPrice,
			&lot.CurrentPrice, &lot.EndTime, &lot.Status)
		if err != nil {
			continue
		}

		lots = append(lots, map[string]interface{}{
			"id":            lot.ID,
			"title":         lot.Title,
			"description":   lot.Description,
			"start_price":   lot.StartPrice,
			"current_price": lot.CurrentPrice,
			"end_time":      lot.EndTime,
			"status":        lot.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lots)
}

// CreateLot - создание нового лота
func (h *LotHandler) CreateLot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method nod supported", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		StartPrice  int    `json:"start_price"`
		EndTime     string `json:"end_time"`
		UserID      int    `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid data", http.StatusBadRequest)
		return
	}
	//"2024-06-15T14:30:00+03:00"
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "Invalid time format", http.StatusBadRequest)
		return
	}

	// Сохранение в базу данных
	var lotID int
	createdAt := time.Now()
	err = h.DB.QueryRow(
		`INSERT INTO lots (title, description, start_price, current_price, end_time, user_id, status, created_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		req.Title, req.Description, req.StartPrice, req.StartPrice, endTime,
		req.UserID, "active", createdAt,
	).Scan(&lotID)

	if err != nil {
		fmt.Printf(err.Error())
		http.Error(w, "error creating lot", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "lot created successfully",
		"lot_id":  lotID,
	})
}

// GetLot - получение информации о конкретном лоте
func (h *LotHandler) GetLot(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID из URL
	idStr := r.URL.Path[len("/api/lots/"):]
	lotID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid lot ID", http.StatusBadRequest)
		return
	}

	var lot struct {
		ID           int       `json:"id"`
		Title        string    `json:"title"`
		Description  string    `json:"description"`
		StartPrice   float64   `json:"start_price"`
		CurrentPrice float64   `json:"current_price"`
		EndTime      time.Time `json:"end_time"`
		Status       string    `json:"status"`
	}

	err = h.DB.QueryRow(`
		SELECT id, title, description, start_price, current_price, end_time, status 
		FROM lots WHERE id = $1
	`, lotID).Scan(&lot.ID, &lot.Title, &lot.Description, &lot.StartPrice,
		&lot.CurrentPrice, &lot.EndTime, &lot.Status)

	if err != nil {
		http.Error(w, "lot not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lot)
}
