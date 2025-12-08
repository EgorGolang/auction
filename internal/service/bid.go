package service

import (
	"auction/internal/errs"
	"auction/internal/models"
	"auction/internal/repository"
	"context"
	"fmt"
)

type BidService struct {
	bidRepo repository.BidRepository
	lotRepo repository.LotRepository
}

func NewBidService(bidRepo repository.BidRepository, lotRepo repository.LotRepository) *BidService {
	return &BidService{
		bidRepo: bidRepo,
		lotRepo: lotRepo,
	}
}

func (s *BidService) CreateBid(ctx context.Context, userID int, bid models.PlaceBid) (*models.BidResponse, error) {
	lot, err := s.lotRepo.GetLotByID(ctx, bid.LotID)
	if err != nil {
		return nil, err
	}
	if lot == nil {
		return nil, fmt.Errorf("lot %d not found", errs.ErrFoundLot)
	}
	if bid.Amount <= lot.CurrentPrice {
		return nil, errs.ErrBidTooLow
	}
	if lot.UserID == userID {
		return nil, errs.ErrCannotBidOnOwnLot
	}
	bidData := models.BidCreate{
		LotID:  bid.LotID,
		UserID: userID,
		Amount: bid.Amount,
	}

	bidID, err := s.bidRepo.CreateBid(ctx, bidData)
	if err != nil {
		return nil, err
	}
	err = s.lotRepo.UpdateLotPrice(ctx, bid.LotID, bid.Amount)
	if err != nil {
		return nil, err
	}
	return &models.BidResponse{
		ID:     bidID,
		LotID:  bid.LotID,
		UserID: userID,
		Amount: bid.Amount,
	}, nil
}

func (s *BidService) GetMyBids(ctx context.Context, userID int) ([]models.Bid, error) {
	bidInfos, err := s.bidRepo.GetMyBids(ctx, userID)
	if err != nil {
		return nil, err
	}
	var bids []models.Bid
	for _, bidInfo := range bidInfos {
		bid := models.Bid{
			ID:        bidInfo.ID,
			LotID:     bidInfo.LotID,
			UserID:    userID,
			Amount:    bidInfo.Amount,
			CreatedAt: bidInfo.CreatedAt,
		}
		bids = append(
			bids,
			bid,
		)
	}
	return bids, nil
}
