package handlers

import (
	"auction/internal/errs"
	"auction/internal/middleware"
	"auction/internal/models"
	"auction/internal/service"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type BidHandler struct {
	db         *sql.DB
	bidService *service.BidService
}

func NewBidHandler(db *sql.DB, bidService *service.BidService) *BidHandler {
	return &BidHandler{
		db:         db,
		bidService: bidService,
	}
}

// @Summary Создание ставки на лот
// @Description Позволяет пользователю сделать ставку на активный лот
// @Tags bids
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.PlaceBidRequest true "Данные для создания ставки"
// @Success 201 {object} models.BidResponse "Ставка успешно создана"
// @Failure 400 {object} models.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} models.ErrorResponse "Не авторизован"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен (не пользователь)"
// @Failure 404 {object} models.ErrorResponse "Лот не найден"
// @Failure 409 {object} models.ErrorResponse "Нельзя делать ставку на собственный лот"
// @Failure 422 {object} models.ErrorResponse "Ставка ниже текущей цены"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/bids/create [post]
func (h *BidHandler) CreateBid(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if user.Role != "user" {
		http.Error(w, "no access", http.StatusUnauthorized)
		return
	}
	var bid models.PlaceBid
	if err := json.NewDecoder(r.Body).Decode(&bid); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	bidResponse, err := h.bidService.CreateBid(r.Context(), user.ID, bid)
	if err != nil {
		switch err {
		case errs.ErrFoundLot:
			http.Error(w, "lot not found", http.StatusNotFound)
		case errs.ErrBidTooLow:
			http.Error(w, "lot too low", http.StatusBadRequest)
		case errs.ErrCannotBidOnOwnLot:
			http.Error(w, "cannot bid on own lot", http.StatusBadRequest)
		default:
			log.Printf("error creating bid: %v", err)
			http.Error(w, "error creating bid", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bidResponse)
}

// @Summary Получение всех ставок пользователя
// @Description Возвращает список всех ставок, сделанных текущим пользователем
// @Tags bids
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserBidsResponse "Список ставок пользователя"
// @Failure 401 {object} models.ErrorResponse "Не авторизован"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен (не пользователь)"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/bids/my [get]
func (h *BidHandler) GetMyBids(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if user.Role != "user" {
		http.Error(w, "no access", http.StatusUnauthorized)
		return
	}
	bids, err := h.bidService.GetMyBids(r.Context(), user.ID)
	if err != nil {
		log.Printf("error getting bids: %v", err)
		http.Error(w, "error getting bids", http.StatusInternalServerError)
		return
	}
	response := models.UserBidsResponse{
		UserID: user.ID,
		Bids:   bids,
		Count:  len(bids),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
