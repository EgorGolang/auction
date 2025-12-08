package handlers

import (
	"auction/internal/models"
	"auction/internal/pkg"
	"auction/internal/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{
		db: db,
	}
}

// @Summary Регистрация нового пользователя
// @Description Создание нового пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.SignInRequest true "Данные для регистрации"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var existingEmail string
	err := h.db.QueryRow("SELECT email FROM users WHERE email = $1", req.Email).Scan(&existingEmail)
	if err != sql.ErrNoRows {
		http.Error(w, "email already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to process password", http.StatusInternalServerError)
		return
	}

	var userID int
	err = h.db.QueryRow("INSERT INTO users(username, email, password_hash, role) VALUES ($1,$2,$3,$4) RETURNING id",
		req.Username,
		req.Email,
		hashedPassword,
		"user").Scan(&userID)

	if err != nil {
		log.Printf("database INSERT error: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
		Role:     "user",
	}

	accessToken, refreshToken, err := h.generateTokenPair(user)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// @Summary Авторизация пользователя
// @Description Авторизует пользователя и возвращает токены
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.SignInRequest true "Данные для авторизации"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	var hashedPassword string
	err := h.db.QueryRow(
		"SELECT id, username, email, password_hash, role FROM users WHERE username = $1",
		req.Username,
	).Scan(&user.ID,
		&user.Username,
		&user.Email,
		&hashedPassword,
		&user.Role)

	if !utils.CheckPasswordHash(req.Password, hashedPassword) {
		log.Printf("password mismatch for user: %v", err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, refreshToken, err := h.generateTokenPair(user)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := pkg.ValidateToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	if claims.TokenType != "refresh" {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var user models.User
	err = h.db.QueryRow("SELECT id, username, email, role FROM users WHERE id = $1").Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role)

	if err != nil {
		log.Printf("database INSERT error: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	accessToken, refreshToken, err := h.generateTokenPair(user)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) generateTokenPair(user models.User) (accessToken string, refreshToken string, err error) {
	accessToken, err = pkg.GenerateToken(user.ID, user.Username, user.Email, user.Role, "access")
	if err != nil {
		return "", "", errors.New("error generating token")
	}

	refreshToken, err = pkg.GenerateToken(user.ID, user.Username, user.Email, user.Role, "refresh")
	if err != nil {
		return "", "", errors.New("error generating token")
	}
	return accessToken, refreshToken, nil
}
