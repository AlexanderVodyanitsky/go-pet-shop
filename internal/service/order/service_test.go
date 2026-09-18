package order

import (
	"context"
	"go-pet-shop/internal/models"
	"testing"
)

type repositoryStub struct {
	getOrderByIDFunc           func(context.Context, int) (models.Order, error)
	getOrderItemsByOrderIDFunc func(context.Context, int) ([]models.OrderItem, error)
}

func (r *repositoryStub) CreateOrder(context.Context, models.Order) (int, error) {
	return 0, nil
}

func (r *repositoryStub) AddOrderItem(context.Context, models.OrderItem) error {
	return nil
}

func (r *repositoryStub) GetOrderByID(ctx context.Context, id int) (models.Order, error) {
	return r.getOrderByIDFunc(ctx, id)
}

func (r *repositoryStub) GetOrdersByUserEmail(context.Context, string) ([]models.Order, error) {
	return nil, nil
}

func (r *repositoryStub) GetOrderItemsByOrderID(
	ctx context.Context,
	id int,
) ([]models.OrderItem, error) {
	return r.getOrderItemsByOrderIDFunc(ctx, id)
}

func TestGetOrderByIDAggregatesItems(t *testing.T) {
	repository := &repositoryStub{
		getOrderByIDFunc: func(_ context.Context, id int) (models.Order, error) {
			return models.Order{ID: id, UserEmail: "alex@example.com", TotalPrice: 20}, nil
		},
		getOrderItemsByOrderIDFunc: func(_ context.Context, id int) ([]models.OrderItem, error) {
			return []models.OrderItem{{ID: 1, OrderID: id, ProductID: 2, Quantity: 3}}, nil
		},
	}

	order, err := New(repository).GetOrderByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if len(order.Items) != 1 || order.Items[0].OrderID != 7 {
		t.Fatalf("order was not aggregated: %+v", order)
	}
}
