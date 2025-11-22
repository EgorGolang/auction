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

type BidHandler struct {
	db *sql.DB
}

func NewBidHandler(db *sql.DB) *BidHandler {
	return &BidHandler{db: db}
}

// CreateBid - создание ставки
func (h *BidHandler) CreateBid(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var bid models.Bid

	if err := json.NewDecoder(r.Body).Decode(&bid); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	bid.UserID = user.ID

	var currentPrice float64
	err := h.db.QueryRow(
		"SELECT current_price FROM lots WHERE id = $1",
		bid.LotID).Scan(&currentPrice)
	if err == sql.ErrNoRows {
		http.Error(w, "lot not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "lot not found", http.StatusNotFound)
		return
	}

	/*if bid.Amount <= currentPrice {
		http.Error(w, "Bid must be higher than current price", http.StatusBadRequest)
		return
	}*/
	var bidID int
	err = h.db.QueryRow(`INSERT INTO bids (lot_id, user_id, amount) VALUES ($1, $2, $3) RETURNING id`,
		bid.LotID,
		bid.UserID,
		bid.Amount).Scan(&bidID)

	if err != nil {
		http.Error(w, "error created bid", http.StatusNotFound)
		return
	}

	_, err = h.db.Exec(
		"UPDATE lots SET current_price = $1 WHERE id = $2",
		bid.Amount, bid.LotID)
	if err != nil {
		http.Error(w, "error updated bid", http.StatusNotFound)
		return
	}

	bid.ID = bidID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bid)
}

func (h *BidHandler) GetMyBids(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "Не указан user_id", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный формат user_id", http.StatusBadRequest)
		return
	}

	rows, err := h.db.Query(`
        SELECT
            b.id,
            b.lot_id,
            l.title as lot_title,
            b.amount,
            b.created_at,
            l.current_price,
            CASE
                WHEN b.amount = l.current_price THEN 'winning'
                ELSE 'outbid'
            END as bid_status
        FROM bids b
        JOIN lots l ON b.lot_id = l.id
        WHERE b.user_id = $1
        ORDER BY b.created_at DESC
    `, userID)

	if err != nil {
		log.Printf("Ошибка базы данных: %v", err)
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var bids []map[string]interface{}
	for rows.Next() {
		var bid struct {
			ID           int       `json:"id"`
			LotID        int       `json:"lot_id"`
			LotTitle     string    `json:"lot_title"`
			Amount       float64   `json:"amount"`
			CreatedAt    time.Time `json:"created_at"`
			CurrentPrice float64   `json:"current_price"`
			BidStatus    string    `json:"bid_status"`
		}

		err := rows.Scan(&bid.ID, &bid.LotID, &bid.LotTitle, &bid.Amount,
			&bid.CreatedAt, &bid.CurrentPrice, &bid.BidStatus)
		if err != nil {
			log.Printf("Ошибка сканирования ставки: %v", err)
			continue
		}

		bids = append(bids, map[string]interface{}{
			"id":            bid.ID,
			"lot_id":        bid.LotID,
			"lot_title":     bid.LotTitle,
			"amount":        bid.Amount,
			"created_at":    bid.CreatedAt,
			"current_price": bid.CurrentPrice,
			"bid_status":    bid.BidStatus, //
		})
	}

	if err = rows.Err(); err != nil {
		log.Printf("Ошибка при итерации по ставкам: %v", err)
		http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"bids":    bids,
		"count":   len(bids),
	})
}
