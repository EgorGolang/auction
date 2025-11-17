package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

// Register - регистрация пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not found", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	// Сохранение в базу данных
	_, err = h.DB.Exec(
		"INSERT INTO users (username, email, password_hash, created_at) VALUES ($1, $2, $3, $4)",
		req.Username, req.Email, string(hashedPassword), time.Now(),
	)

	if err != nil {
		http.Error(w, "user already exists", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "user registered successfully",
	})
}

// Login - вход в систему
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method is not supported", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid data", http.StatusBadRequest)
		return
	}

	// Поиск пользователя в базе
	var userID int
	var storedPassword string
	err := h.DB.QueryRow(
		"SELECT id, password_hash FROM users WHERE username = $1",
		req.Username,
	).Scan(&userID, &storedPassword)

	if err != nil {
		http.Error(w, "incorrect login or password", http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password)); err != nil {
		http.Error(w, "incorrect login or password", http.StatusUnauthorized)
		return
	}

	// Здесь можно добавить JWT токен или сессию

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "successfully logged in",
		"user_id":  userID,
		"username": req.Username,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "successful exit",
	})
}
