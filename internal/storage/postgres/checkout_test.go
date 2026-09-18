package postgres

import (
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"testing"
)

func TestCombineOrderItems(t *testing.T) {
	items, err := combineOrderItems([]models.OrderItem{
		{ProductID: 4, Quantity: 2},
		{ProductID: 2, Quantity: 1},
		{ProductID: 4, Quantity: 3},
	})
	if err != nil {
		t.Fatalf("combine items: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ProductID != 2 || items[0].Quantity != 1 {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[1].ProductID != 4 || items[1].Quantity != 5 {
		t.Fatalf("unexpected second item: %+v", items[1])
	}
}

func TestCombineOrderItemsRejectsInvalidQuantity(t *testing.T) {
	_, err := combineOrderItems([]models.OrderItem{{ProductID: 1, Quantity: 0}})
	if !errors.Is(err, storage.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}
