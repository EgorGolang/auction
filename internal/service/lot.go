package service

import (
	"auction/internal/errs"
	"auction/internal/models"
	"auction/internal/repository"
	"context"
	"time"
)

type LotService struct {
	lotRepo  repository.LotRepository
	userRepo repository.UserRepository
}

func NewLotService(lotRepo *repository.PostgresLotRepository, userRepo repository.UserRepository) *LotService {
	return &LotService{
		lotRepo:  lotRepo,
		userRepo: userRepo,
	}
}

func (s *LotService) CreateLot(ctx context.Context, userID int, lot models.Lot) (lotID int, err error) {
	role, err := s.userRepo.GetUserRole(ctx, userID)
	if err != nil {
		return 0, err
	}

	if role != "user" {
		return 0, errs.ErrNoAccess
	}

	if err := s.validateLot(lot); err != nil {
		return 0, err
	}

	lotData := models.LotCreate{
		Title:        lot.Title,
		Description:  lot.Description,
		StartPrice:   lot.StartPrice,
		CurrentPrice: lot.StartPrice,
		EndTime:      lot.EndTime,
		UserID:       userID,
		CreatedAt:    time.Now(),
	}

	lotID, err = s.lotRepo.CreateLot(ctx, lotData)
	if err != nil {
		return 0, err
	}

	return lotID, nil
}

func (s *LotService) GetLots(ctx context.Context) ([]models.LotResponse, error) {
	lots, err := s.lotRepo.GetLots(ctx)
	if err != nil {
		return nil, err
	}
	return lots, nil
}

func (s *LotService) GetLotByID(ctx context.Context, lotID int) (*models.LotResponse, error) {
	if lotID <= 0 {
		return nil, errs.ErrInvalidLotID
	}
	lot, err := s.lotRepo.GetLotByID(ctx, lotID)
	if err != nil {
		return nil, err
	}
	return lot, nil
}

func (s *LotService) validateLot(lot models.Lot) error {
	now := time.Now()
	if len(lot.Title) < 3 {
		return errs.ErrInvalidTitle
	}

	if len(lot.Description) < 10 {
		return errs.ErrInvalidDescription
	}

	if lot.StartPrice <= 0 {
		return errs.ErrInvalidPrice
	}

	if lot.EndTime.Before(now) {
		return errs.ErrEmptyEndTime
	}

	return nil
}

func (s *LotService) DeleteLot(ctx context.Context, lotID int, userID int) error {
	if lotID <= 0 {
		return errs.ErrInvalidLotID
	}
	userRole, err := s.userRepo.GetUserRole(ctx, userID)
	if err != nil {
		return err
	}
	if userRole != "admin" {
		return errs.ErrAdminAccessDenied
	}
	return s.lotRepo.DeleteLot(ctx, lotID)
}
