package repository

import (
	"auction/internal/models"
	"context"
	"database/sql"
)

type BidRepository interface {
	CreateBid(ctx context.Context, bid models.BidCreate) (int, error)
	GetMyBids(ctx context.Context, userID int) ([]models.Bid, error)
}

type PostgresBidRepository struct {
	db      *sql.DB
	bidRepo *BidRepository
}

func NewPostgresBidRepository(db *sql.DB) *PostgresBidRepository {
	return &PostgresBidRepository{db: db}
}

func (r *PostgresBidRepository) CreateBid(ctx context.Context, bid models.BidCreate) (int, error) {
	var bidID int
	err := r.db.QueryRowContext(ctx, "INSERT INTO bids (lot_id, user_id, amount) VALUES ($1, $2, $3) "+
		"RETURNING id", bid.LotID, bid.UserID, bid.Amount).Scan(&bidID)
	if err != nil {
		return 0, err
	}
	return bidID, nil
}

func (r *PostgresBidRepository) GetMyBids(ctx context.Context, userID int) ([]models.Bid, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, lot_id, user_id, amount, created_at FROM bids WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bids []models.Bid
	for rows.Next() {
		var bid models.Bid
		err := rows.Scan(&bid.ID, &bid.LotID, &bid.UserID, &bid.Amount, &bid.CreatedAt)
		if err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bids, nil
}
