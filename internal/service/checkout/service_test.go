package checkout

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	serviceerrors "go-pet-shop/internal/service"
	"testing"
)

type repositoryStub struct {
	placeOrderFunc func(context.Context, string, []models.OrderItem) (int, error)
}

func (r *repositoryStub) PlaceOrder(
	ctx context.Context,
	email string,
	items []models.OrderItem,
) (int, error) {
	return r.placeOrderFunc(ctx, email, items)
}

func TestPlaceOrderNormalizesEmail(t *testing.T) {
	repository := &repositoryStub{
		placeOrderFunc: func(_ context.Context, email string, _ []models.OrderItem) (int, error) {
			if email != "alex@example.com" {
				t.Fatalf("unexpected normalized email: %s", email)
			}
			return 3, nil
		},
	}

	id, err := New(repository).PlaceOrder(context.Background(), " Alex@Example.com ", []models.OrderItem{
		{ProductID: 1, Quantity: 2},
	})
	if err != nil || id != 3 {
		t.Fatalf("unexpected result: id=%d err=%v", id, err)
	}
}

func TestPlaceOrderRejectsEmptyItemsBeforeRepository(t *testing.T) {
	called := false
	repository := &repositoryStub{
		placeOrderFunc: func(context.Context, string, []models.OrderItem) (int, error) {
			called = true
			return 0, nil
		},
	}

	_, err := New(repository).PlaceOrder(context.Background(), "alex@example.com", nil)
	if !errors.Is(err, serviceerrors.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if called {
		t.Fatal("repository must not be called for empty checkout")
	}
}
