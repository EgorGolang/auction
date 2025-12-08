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
	Amount int `json:"amount"`
}

type PlaceBid struct {
	ID     int `json:"id"`
	Amount int `json:"amount"`
	LotID  int `json:"lot_id"`
}

type BidCreate struct {
	LotID  int `json:"lot_id"`
	UserID int `json:"user_id"`
	Amount int `json:"amount"`
}

type BidResponse struct {
	ID     int `json:"id"`
	LotID  int `json:"lot_id"`
	UserID int `json:"user_id"`
	Amount int `json:"amount"`
}

type UserBidsResponse struct {
	UserID int   `json:"user_id"`
	Bids   []Bid `json:"bids"`
	Count  int   `json:"count"`
}
