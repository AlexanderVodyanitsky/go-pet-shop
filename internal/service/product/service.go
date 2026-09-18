package product

import (
	"context"
	"go-pet-shop/internal/models"
	serviceerrors "go-pet-shop/internal/service"
	"strings"
)

type Repository interface {
	GetAllProducts(ctx context.Context) ([]models.Product, error)
	GetProductByID(ctx context.Context, id int) (models.Product, error)
	CreateProduct(ctx context.Context, product models.Product) (int, error)
	DeleteProduct(ctx context.Context, id int) error
	UpdateProduct(ctx context.Context, product models.Product) error
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	products, err := s.repository.GetAllProducts(ctx)
	return products, serviceerrors.MapStorageError(err)
}

func (s *Service) GetProductByID(ctx context.Context, id int) (models.Product, error) {
	if id < 1 {
		return models.Product{}, serviceerrors.InvalidInput("product ID must be a positive integer")
	}

	product, err := s.repository.GetProductByID(ctx, id)
	return product, serviceerrors.MapStorageError(err)
}

func (s *Service) CreateProduct(ctx context.Context, product models.Product) (models.Product, error) {
	if err := validateProduct(&product); err != nil {
		return models.Product{}, err
	}

	id, err := s.repository.CreateProduct(ctx, product)
	if err != nil {
		return models.Product{}, serviceerrors.MapStorageError(err)
	}
	product.ID = id

	return product, nil
}

func (s *Service) UpdateProduct(
	ctx context.Context,
	id int,
	product models.Product,
) (models.Product, error) {
	if id < 1 {
		return models.Product{}, serviceerrors.InvalidInput("product ID must be a positive integer")
	}
	product.ID = id
	if err := validateProduct(&product); err != nil {
		return models.Product{}, err
	}

	if err := s.repository.UpdateProduct(ctx, product); err != nil {
		return models.Product{}, serviceerrors.MapStorageError(err)
	}

	return product, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id int) error {
	if id < 1 {
		return serviceerrors.InvalidInput("product ID must be a positive integer")
	}
	return serviceerrors.MapStorageError(s.repository.DeleteProduct(ctx, id))
}

func validateProduct(product *models.Product) error {
	product.Name = strings.TrimSpace(product.Name)
	if product.Name == "" {
		return serviceerrors.InvalidInput("product name is required")
	}
	if product.Price < 0 {
		return serviceerrors.InvalidInput("product price cannot be negative")
	}
	if product.Stock < 0 {
		return serviceerrors.InvalidInput("product stock cannot be negative")
	}
	return nil
}
