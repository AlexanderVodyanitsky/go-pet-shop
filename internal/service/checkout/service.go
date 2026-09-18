package checkout

import (
	"context"
	"go-pet-shop/internal/models"
	serviceerrors "go-pet-shop/internal/service"
)

type Repository interface {
	PlaceOrder(ctx context.Context, userEmail string, items []models.OrderItem) (int, error)
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) PlaceOrder(
	ctx context.Context,
	userEmail string,
	items []models.OrderItem,
) (int, error) {
	email, err := serviceerrors.NormalizeEmail(userEmail)
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, serviceerrors.InvalidInput("at least one order item is required")
	}
	for _, item := range items {
		if item.ProductID < 1 {
			return 0, serviceerrors.InvalidInput("product ID must be a positive integer")
		}
		if item.Quantity < 1 {
			return 0, serviceerrors.InvalidInput("quantity must be a positive integer")
		}
	}

	orderID, err := s.repository.PlaceOrder(ctx, email, items)
	return orderID, serviceerrors.MapStorageError(err)
}
