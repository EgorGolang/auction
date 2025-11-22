package handlers

import (
	"auction/internal/models"
	"auction/internal/pkg"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingEmail string
	err := h.db.QueryRow("SELECT email FROM users WHERE email = $1", req.Email).Scan(&existingEmail)
	if err != sql.ErrNoRows {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := pkg.HashPassword(req.Password)
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
		log.Printf("Database INSERT error: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
		Role:     "user",
	}

	token, err := pkg.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}

	response := models.AuthRequest{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	var req models.LoginRequest

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

	if !pkg.CheckPasswordHash(req.Password, hashedPassword) {
		log.Printf("Password mismatch for user: %v", err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := pkg.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	response := models.AuthRequest{
		Token: token,
		User:  user,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
