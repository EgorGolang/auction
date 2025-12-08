package models

import (
	"time"
)

type LotCreate struct {
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	StartPrice   int       `json:"start_price"`
	CurrentPrice int       `json:"current_price"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       int       `json:"user_id"`
	EndTime      time.Time `json:"end_time"`
}

type CreateLotRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartPrice  int       `json:"start_price"`
	EndTime     time.Time `json:"end_time"`
}

type Lot struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartPrice  int       `json:"start_price"`
	EndTime     time.Time `json:"end_time"`
}

type LotResponse struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	StartPrice   int       `json:"start_price"`
	CurrentPrice int       `json:"current_price"`
	EndTime      time.Time `json:"end_time"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       int       `json:"user_id"`
}
type CreateLotResponse struct {
	Message string `json:"message"`
	LotID   int    `json:"lot_id"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}
