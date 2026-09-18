package analytics

import (
	"context"
	"go-pet-shop/internal/models"
	serviceerrors "go-pet-shop/internal/service"
)

type Repository interface {
	GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error)
	GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error)
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetUserOrderHistory(
	ctx context.Context,
	email string,
) ([]models.OrderDetail, error) {
	normalizedEmail, err := serviceerrors.NormalizeEmail(email)
	if err != nil {
		return nil, err
	}

	history, err := s.repository.GetUserOrderHistory(ctx, normalizedEmail)
	return history, serviceerrors.MapStorageError(err)
}

func (s *Service) GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error) {
	products, err := s.repository.GetPopularProducts(ctx)
	return products, serviceerrors.MapStorageError(err)
}
