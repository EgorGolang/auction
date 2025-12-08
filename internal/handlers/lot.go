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
	"strconv"
)

type LotHandler struct {
	db         *sql.DB
	lotService *service.LotService
}

func NewLotHandler(db *sql.DB, lotService *service.LotService) *LotHandler {
	return &LotHandler{
		db:         db,
		lotService: lotService}
}

// @Summary Получение списка лотов
// @Description Возвращает список активных лотов
// @Tags lots
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/lots [get]
func (h *LotHandler) GetLots(w http.ResponseWriter, r *http.Request) {
	lots, err := h.lotService.GetLots(r.Context())
	if err != nil {
		log.Println("getLots: ", err)
		return
	}
	if lots == nil {
		lots = []models.LotResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lots)
}

// @Summary Создание нового лота
// @Description Создает новый лот на аукционе
// @Tags lots
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Lot true "Данные лота"
// @Success 201 {object} models.CreateLotResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/lots/create [post]
func (h *LotHandler) CreateLot(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var lot models.Lot
	if err := json.NewDecoder(r.Body).Decode(&lot); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	lotID, err := h.lotService.CreateLot(r.Context(), user.ID, lot)
	if err != nil {
		switch err {
		case errs.ErrNoAccess:
			http.Error(w, "access denied", http.StatusUnauthorized)
		case errs.ErrInvalidTimeFormat:
			http.Error(w, "invalid time format", http.StatusBadRequest)
		case errs.ErrInvalidTitle:
			http.Error(w, "invalid title", http.StatusBadRequest)
		case errs.ErrInvalidDescription:
			http.Error(w, "invalid description", http.StatusBadRequest)
		case errs.ErrInvalidPrice:
			http.Error(w, "invalid price", http.StatusBadRequest)
		default:
			log.Printf("error creating lot: %v", err)
			http.Error(w, "error creating lot", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "lot created successfully",
		"lot_id":  lotID,
	})
}

// @Summary Получение лота по ID
// @Description Возвращает информацию о конкретном лоте по его ID
// @Tags lots
// @Accept json
// @Produce json
// @Param id query int true "ID лота" minimum(1)
// @Success 200 {object} models.LotResponse "Информация о лоте"
// @Failure 400 {object} models.ErrorResponse "Неверный ID"
// @Failure 404 {object} models.ErrorResponse "Лот не найден"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/lot [get]
func (h *LotHandler) GetLotByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "no id provided", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.Error(w, "invalid lot ID", http.StatusBadRequest)
		return
	}
	lot, err := h.lotService.GetLotByID(r.Context(), id)
	if err != nil {
		switch err {
		case errs.ErrFoundLot:
			http.Error(w, "lot not found", http.StatusNotFound)
		case errs.ErrInvalidLotID:
			http.Error(w, "invalid lot ID", http.StatusBadRequest)
		default:
			log.Printf("error getting lot: %v", err)
			http.Error(w, "error getting lot", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lot)
}

// @Summary Удаление лота
// @Description Удаляет лот по ID (только для администраторов)
// @Tags lots
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query int true "ID лота для удаления" minimum(1)
// @Success 204 "Лот успешно удален"
// @Failure 400 {object} models.ErrorResponse "Неверный ID или отсутствует ID"
// @Failure 401 {object} models.ErrorResponse "Не авторизован"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен (не администратор)"
// @Failure 404 {object} models.ErrorResponse "Лот не найден"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/lot/delete [delete]
func (h *LotHandler) DeleteLot(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		log.Println("deleteLot: user is nil")
		http.Error(w, "admin access required", http.StatusUnauthorized)
		return
	}

	if user.Role != "admin" {
		log.Printf("deleteLot: access denied. Role '%s' != 'admin'", user.Role)
		http.Error(w, "admin access required", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "lot ID is required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.Error(w, "invalid lot ID", http.StatusBadRequest)
		return
	}
	err = h.lotService.DeleteLot(r.Context(), id, user.ID)
	if err != nil {
		switch err {
		case errs.ErrAdminAccessDenied:
			http.Error(w, "access denied", http.StatusForbidden)
		case errs.ErrFoundLot:
			http.Error(w, "lot not found", http.StatusNotFound)
		default:
			log.Printf("error deleting lot: %v", err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
