package models

import (
	"time"
)

type Lot struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	StartPrice   int       `json:"start_price"`
	CurrentPrice int       `json:"current_price"`
	Step         int       `json:"step"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	EndTime      string    `json:"end_time"`
	UserID       string    `json:"user_id"`
}

type CreateLotRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartPrice  int       `json:"start_price"`
	Step        int       `json:"step"`
	EndTime     time.Time `json:"end_time"`
}
