package storage

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrConflict          = errors.New("conflict")
	ErrInsufficientStock = errors.New("insufficient stock")
)
