package analytics

import (
	"context"
	"go-pet-shop/internal/models"
	"testing"
)

type repositoryStub struct {
	getHistoryFunc func(context.Context, string) ([]models.OrderDetail, error)
}

func (r *repositoryStub) GetUserOrderHistory(
	ctx context.Context,
	email string,
) ([]models.OrderDetail, error) {
	return r.getHistoryFunc(ctx, email)
}

func (r *repositoryStub) GetPopularProducts(context.Context) ([]models.PopularProduct, error) {
	return nil, nil
}

func TestGetUserOrderHistoryNormalizesEmail(t *testing.T) {
	repository := &repositoryStub{
		getHistoryFunc: func(_ context.Context, email string) ([]models.OrderDetail, error) {
			if email != "alex@example.com" {
				t.Fatalf("unexpected normalized email: %s", email)
			}
			return []models.OrderDetail{}, nil
		},
	}

	_, err := New(repository).GetUserOrderHistory(context.Background(), " Alex@Example.com ")
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
}
