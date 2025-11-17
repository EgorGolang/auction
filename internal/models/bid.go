package models

import "time"

type Bid struct {
	ID        int       `json:"id"`
	LotID     int       `json:"lot_id"`
	UserID    int       `json:"user_id"`
	Amount    int       `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type PlaceBidRequest struct {
	LotID  int `json:"lot_id"`
	UserID int `json:"user_id"`
	Amount int `json:"amount"`
}
