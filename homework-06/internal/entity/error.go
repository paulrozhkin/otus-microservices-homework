package entity

import "errors"

var (
	ErrAlreadyExists      = errors.New("already exists")
	ErrNotFound           = errors.New("not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrServiceUnavailable = errors.New("service unavailable")
)

type ErrorResponse struct {
	Code    int32  `json:"code"  binding:"required"`
	Message string `json:"message" binding:"required"`
}
