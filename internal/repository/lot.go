package repository

import (
	"auction/internal/errs"
	"auction/internal/models"
	"context"
	"database/sql"
)

type LotRepository interface {
	CreateLot(ctx context.Context, lot models.LotCreate) (int, error)
	GetLots(ctx context.Context) ([]models.LotResponse, error)
	GetLotByID(ctx context.Context, id int) (*models.LotResponse, error)
	DeleteLot(ctx context.Context, id int) error
	UpdateLotPrice(ctx context.Context, lotID int, newPrice int) error
}

type UserRepository interface {
	GetUserRole(ctx context.Context, userID int) (string, error)
}

type PostgresLotRepository struct {
	db      *sql.DB
	lotRepo *LotRepository
}

func NewPostgresLotRepository(db *sql.DB) *PostgresLotRepository {
	return &PostgresLotRepository{db: db}
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresLotRepository) CreateLot(ctx context.Context, lot models.LotCreate) (int, error) {
	var lotID int
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO lots (title, description, start_price, current_price, end_time, user_id, created_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		lot.Title, lot.Description, lot.StartPrice, lot.CurrentPrice, lot.EndTime, lot.UserID, lot.CreatedAt,
	).Scan(&lotID)

	if err != nil {
		return 0, err
	}

	return lotID, nil
}

func (r *PostgresLotRepository) GetLots(ctx context.Context) ([]models.LotResponse, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, title, description, start_price, current_price, end_time, 
       created_at, user_id FROM lots WHERE status = 'active' ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lots []models.LotResponse
	for rows.Next() {
		var lot models.LotResponse
		err := rows.Scan(
			&lot.ID,
			&lot.Title,
			&lot.Description,
			&lot.StartPrice,
			&lot.CurrentPrice,
			&lot.EndTime,
			&lot.CreatedAt,
			&lot.UserID)
		if err != nil {
			return nil, err
		}
		lots = append(lots, lot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return lots, nil
}

func (r *PostgresLotRepository) GetLotByID(ctx context.Context, id int) (*models.LotResponse, error) {
	if id <= 0 {
		return nil, errs.ErrFoundLot
	}
	query := `SELECT id, title, description, start_price, current_price, end_time FROM lots WHERE id = $1`
	lot := &models.LotResponse{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lot.ID,
		&lot.Title,
		&lot.Description,
		&lot.StartPrice,
		&lot.CurrentPrice,
		&lot.EndTime)
	if err == sql.ErrNoRows {
		return nil, errs.ErrFoundLot
	}
	if err != nil {
		return nil, err
	}

	return lot, nil
}

func (r *PostgresLotRepository) DeleteLot(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM lots WHERE id = $1`, id)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresUserRepository) GetUserRole(ctx context.Context, userID int) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx,
		"SELECT role FROM users WHERE id = $1",
		userID,
	).Scan(&role)

	if err != nil {
		return "", err
	}

	return role, nil
}

func (r *PostgresLotRepository) UpdateLotPrice(ctx context.Context, lotID int, newPrice int) error {
	result, err := r.db.ExecContext(ctx, "UPDATE lots SET current_price = $1 WHERE id = $2", newPrice, lotID)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}
