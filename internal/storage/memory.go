package storage

import (
	"auction/internal/models"
	"sync"
	_ "time"
)

type MemoryStorage struct {
	mu    sync.RWMutex
	lots  map[string]*models.Lot
	bids  map[string]*models.Bid
	users map[string]*models.User // Пока упрощенная модель пользователя
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		lots:  make(map[string]*models.Lot),
		bids:  make(map[string]*models.Bid),
		users: make(map[string]*models.User),
	}
}

func (s *MemoryStorage) CreateLot(lot *models.Lot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lots[lot.ID] = lot
	return nil
}

func (s *MemoryStorage) GetLot(id string) (*models.Lot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lot, exists := s.lots[id]
	if !exists {
		return nil, nil
	}
	return lot, nil
}

func (s *MemoryStorage) GetAllLots() ([]*models.Lot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lots := make([]*models.Lot, 0, len(s.lots))
	for _, lot := range s.lots {
		lots = append(lots, lot)
	}
	return lots, nil
}

func (s *MemoryStorage) CreateBid(bid *models.Bid) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.bids[bid.ID] = bid

	// Обновляем текущую цену лота
	if lot, exists := s.lots[bid.LotID]; exists {
		lot.CurrentPrice = bid.Amount
	}

	return nil
}

func (s *MemoryStorage) GetBidsByLot(lotID string) ([]*models.Bid, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bids []*models.Bid
	for _, bid := range s.bids {
		if bid.LotID == lotID {
			bids = append(bids, bid)
		}
	}
	return bids, nil
}

func (s *MemoryStorage) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.ID] = user
	return nil
}

func (s *MemoryStorage) GetUserByID(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (s *MemoryStorage) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (s *MemoryStorage) GetAllUsers() ([]*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users, nil
}
