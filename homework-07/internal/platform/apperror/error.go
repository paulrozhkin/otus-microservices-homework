package apperror

import "errors"

var (
	ErrAlreadyExists      = errors.New("already exists")
	ErrForbidden          = errors.New("forbidden")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidOperation   = errors.New("invalid operation")
	ErrNotFound           = errors.New("not found")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrUnauthorized       = errors.New("unauthorized")
)

type Response struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}
