package service

import (
	"errors"
	"go-pet-shop/internal/storage"
	"testing"
)

func TestNormalizeEmail(t *testing.T) {
	email, err := NormalizeEmail(" Alex@Example.com ")
	if err != nil {
		t.Fatalf("normalize email: %v", err)
	}
	if email != "alex@example.com" {
		t.Fatalf("unexpected email: %s", email)
	}
}

func TestMapStorageError(t *testing.T) {
	err := MapStorageError(storage.ErrNotFound)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected service not found, got %v", err)
	}
}
