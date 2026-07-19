package entity

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidOperation   = errors.New("invalid operation")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrServiceUnavailable = errors.New("service unavailable")
)

type ErrorResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}
