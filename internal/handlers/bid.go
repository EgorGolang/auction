package handlers

import (
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
	if r.Method != http.MethodPost {
		http.Error(w, "method is not supported", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		LotID  int     `json:"lot_id"`
		UserID int     `json:"user_id"` //Добавить JWT
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid data", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли лот и активен ли он
	var currentPrice float64
	var status string
	err := h.db.QueryRow(
		"SELECT current_price, status FROM lots WHERE id = $1",
		req.LotID,
	).Scan(&currentPrice, &status)

	if err != nil {
		http.Error(w, "lot not found", http.StatusNotFound)
		return
	}

	if status != "active" {
		http.Error(w, "lot is not active", http.StatusBadRequest)
		return
	}

	// Проверяем, что ставка выше текущей
	if req.Amount <= currentPrice {
		http.Error(w, "the bid must be higher than the current price", http.StatusBadRequest)
		return
	}

	// Начинаем транзакцию
	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	// Создаем ставку
	_, err = tx.Exec(
		"INSERT INTO bids (lot_id, user_id, amount, created_at) VALUES ($1, $2, $3, $4)",
		req.LotID, req.UserID, req.Amount, time.Now(),
	)

	if err != nil {
		tx.Rollback()
		http.Error(w, "error creating bid", http.StatusInternalServerError)
		return
	}

	// Обновляем текущую цену лота
	_, err = tx.Exec(
		"UPDATE lots SET current_price = $1 WHERE id = $2",
		req.Amount, req.LotID,
	)

	if err != nil {
		tx.Rollback()
		http.Error(w, "price update error", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "bed successfully placed",
	})
}

// GetMyBids - получение ставок текущего пользователя
func (h *BidHandler) GetMyBids(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// В реальном приложении userID должен браться из JWT токена или сессии
	// Сейчас будем использовать query parameter, но это НЕБЕЗОПАСНО - только для тестирования!
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
			"bid_status":    bid.BidStatus, // "winning" - лидируете, "outbid" - вас перебили
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
