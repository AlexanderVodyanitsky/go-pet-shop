package order

import (
	"context"
	"go-pet-shop/internal/models"
	serviceerrors "go-pet-shop/internal/service"
)

type Repository interface {
	CreateOrder(ctx context.Context, order models.Order) (int, error)
	AddOrderItem(ctx context.Context, item models.OrderItem) error
	GetOrderByID(ctx context.Context, id int) (models.Order, error)
	GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error)
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateOrder(ctx context.Context, order models.Order) (models.Order, error) {
	email, err := serviceerrors.NormalizeEmail(order.UserEmail)
	if err != nil {
		return models.Order{}, err
	}
	if order.TotalPrice < 0 {
		return models.Order{}, serviceerrors.InvalidInput("order total price cannot be negative")
	}
	order.UserEmail = email

	id, err := s.repository.CreateOrder(ctx, order)
	if err != nil {
		return models.Order{}, serviceerrors.MapStorageError(err)
	}
	order.ID = id

	return order, nil
}

func (s *Service) AddOrderItem(
	ctx context.Context,
	orderID int,
	item models.OrderItem,
) (models.OrderItem, error) {
	if orderID < 1 {
		return models.OrderItem{}, serviceerrors.InvalidInput("order ID must be a positive integer")
	}
	if item.ProductID < 1 {
		return models.OrderItem{}, serviceerrors.InvalidInput("product ID must be a positive integer")
	}
	if item.Quantity < 1 {
		return models.OrderItem{}, serviceerrors.InvalidInput("quantity must be a positive integer")
	}
	item.OrderID = orderID

	if err := s.repository.AddOrderItem(ctx, item); err != nil {
		return models.OrderItem{}, serviceerrors.MapStorageError(err)
	}

	return item, nil
}

func (s *Service) GetOrderByID(ctx context.Context, id int) (models.Order, error) {
	if id < 1 {
		return models.Order{}, serviceerrors.InvalidInput("order ID must be a positive integer")
	}

	order, err := s.repository.GetOrderByID(ctx, id)
	if err != nil {
		return models.Order{}, serviceerrors.MapStorageError(err)
	}

	items, err := s.repository.GetOrderItemsByOrderID(ctx, id)
	if err != nil {
		return models.Order{}, serviceerrors.MapStorageError(err)
	}
	order.Items = items

	return order, nil
}

func (s *Service) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error) {
	normalizedEmail, err := serviceerrors.NormalizeEmail(email)
	if err != nil {
		return nil, err
	}

	orders, err := s.repository.GetOrdersByUserEmail(ctx, normalizedEmail)
	return orders, serviceerrors.MapStorageError(err)
}
