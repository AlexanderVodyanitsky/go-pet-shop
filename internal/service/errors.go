package service

import (
	"errors"
	"go-pet-shop/internal/storage"
	"net/mail"
	"strings"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrConflict          = errors.New("conflict")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type domainError struct {
	kind    error
	message string
}

func (e *domainError) Error() string {
	return e.message
}

func (e *domainError) Unwrap() error {
	return e.kind
}

func InvalidInput(message string) error {
	return &domainError{kind: ErrInvalidInput, message: message}
}

func MapStorageError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, storage.ErrNotFound):
		return &domainError{kind: ErrNotFound, message: ErrNotFound.Error()}
	case errors.Is(err, storage.ErrInvalidInput):
		return &domainError{kind: ErrInvalidInput, message: ErrInvalidInput.Error()}
	case errors.Is(err, storage.ErrConflict):
		return &domainError{kind: ErrConflict, message: ErrConflict.Error()}
	case errors.Is(err, storage.ErrInsufficientStock):
		return &domainError{kind: ErrInsufficientStock, message: ErrInsufficientStock.Error()}
	default:
		return err
	}
}

func NormalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", InvalidInput("valid user email is required")
	}
	return value, nil
}
