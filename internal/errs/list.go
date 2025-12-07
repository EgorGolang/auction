package errs

import "errors"

var (
	ErrNoAccess              = errors.New("no access")
	ErrInvalidTimeFormat     = errors.New("invalid time format")
	ErrInvalidTitle          = errors.New("title must be at least 3 characters long")
	ErrInvalidDescription    = errors.New("description must be at least 10 characters long")
	ErrInvalidPrice          = errors.New("price must be positive")
	ErrEmptyEndTime          = errors.New("end time is required")
	ErrInvalidLotID          = errors.New("invalid lot ID")
	ErrFoundLot              = errors.New("lot not found")
	ErrAdminAccessDenied     = errors.New("admin access required")
	ErrBidTooLow             = errors.New("bid too low")
	ErrCannotBidOnOwnLot     = errors.New("cannot bid on own lot")
	ErrInvalidUsername       = errors.New("username must be at least 3 characters")
	ErrInvalidPassword       = errors.New("password must be at least 8 characters")
	ErrAlreadyExists         = errors.New("already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
)
