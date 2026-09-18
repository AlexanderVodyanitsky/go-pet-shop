package order

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/service"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
)

type ordersMock struct {
	createOrderFunc          func(context.Context, models.Order) (models.Order, error)
	addOrderItemFunc         func(context.Context, int, models.OrderItem) (models.OrderItem, error)
	getOrderByIDFunc         func(context.Context, int) (models.Order, error)
	getOrdersByUserEmailFunc func(context.Context, string) ([]models.Order, error)
}

func (m *ordersMock) CreateOrder(ctx context.Context, order models.Order) (models.Order, error) {
	if m.createOrderFunc == nil {
		return models.Order{}, nil
	}
	return m.createOrderFunc(ctx, order)
}

func (m *ordersMock) AddOrderItem(
	ctx context.Context,
	orderID int,
	item models.OrderItem,
) (models.OrderItem, error) {
	if m.addOrderItemFunc == nil {
		return models.OrderItem{}, nil
	}
	return m.addOrderItemFunc(ctx, orderID, item)
}

func (m *ordersMock) GetOrderByID(ctx context.Context, id int) (models.Order, error) {
	if m.getOrderByIDFunc == nil {
		return models.Order{}, nil
	}
	return m.getOrderByIDFunc(ctx, id)
}

func (m *ordersMock) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error) {
	if m.getOrdersByUserEmailFunc == nil {
		return nil, nil
	}
	return m.getOrdersByUserEmailFunc(ctx, email)
}

func TestCreateOrder(t *testing.T) {
	mock := &ordersMock{
		createOrderFunc: func(_ context.Context, order models.Order) (models.Order, error) {
			if order.UserEmail != "Alex@Example.com" || order.TotalPrice != 25.5 {
				t.Fatalf("unexpected order: %+v", order)
			}
			return models.Order{ID: 9, UserEmail: "alex@example.com", TotalPrice: 25.5}, nil
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(
		`{"user_email":"Alex@Example.com","total_price":25.5}`,
	))
	w := httptest.NewRecorder()

	New(orderTestLogger(), mock).CreateOrder(w, req)

	assertOrderStatus(t, w, http.StatusCreated)
}

func TestCreateOrderUserNotFound(t *testing.T) {
	mock := &ordersMock{
		createOrderFunc: func(context.Context, models.Order) (models.Order, error) {
			return models.Order{}, service.ErrNotFound
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(
		`{"user_email":"missing@example.com","total_price":0}`,
	))
	w := httptest.NewRecorder()

	New(orderTestLogger(), mock).CreateOrder(w, req)

	assertOrderStatus(t, w, http.StatusNotFound)
}

func TestAddOrderItem(t *testing.T) {
	mock := &ordersMock{
		addOrderItemFunc: func(_ context.Context, orderID int, item models.OrderItem) (models.OrderItem, error) {
			if orderID != 4 || item.ProductID != 2 || item.Quantity != 3 {
				t.Fatalf("unexpected item: %+v", item)
			}
			item.OrderID = orderID
			return item, nil
		},
	}
	req := orderRequestWithID(http.MethodPost, "/orders/4/items", "4",
		`{"product_id":2,"quantity":3}`)
	w := httptest.NewRecorder()

	New(orderTestLogger(), mock).AddOrderItem(w, req)

	assertOrderStatus(t, w, http.StatusCreated)
}

func TestAddOrderItemRejectsZeroQuantity(t *testing.T) {
	mock := &ordersMock{
		addOrderItemFunc: func(context.Context, int, models.OrderItem) (models.OrderItem, error) {
			return models.OrderItem{}, service.InvalidInput("quantity must be a positive integer")
		},
	}
	req := orderRequestWithID(http.MethodPost, "/orders/4/items", "4",
		`{"product_id":2,"quantity":0}`)
	w := httptest.NewRecorder()

	New(orderTestLogger(), mock).AddOrderItem(w, req)

	assertOrderStatus(t, w, http.StatusBadRequest)
}

func TestGetOrderByIDWithItems(t *testing.T) {
	createdAt := time.Date(2026, time.September, 14, 10, 0, 0, 0, time.UTC)
	mock := &ordersMock{
		getOrderByIDFunc: func(_ context.Context, id int) (models.Order, error) {
			return models.Order{
				ID: id, UserEmail: "alex@example.com", TotalPrice: 12, CreatedAt: createdAt,
				Items: []models.OrderItem{{ID: 1, OrderID: id, ProductID: 2, Quantity: 1}},
			}, nil
		},
	}
	req := orderRequestWithID(http.MethodGet, "/orders/5", "5", "")
	w := httptest.NewRecorder()

	New(orderTestLogger(), mock).GetOrderByID(w, req)

	assertOrderStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"items"`) {
		t.Fatalf("expected order items in response: %s", w.Body.String())
	}
}

func TestGetOrderByIDNotFound(t *testing.T) {
	mock := &ordersMock{
		getOrderByIDFunc: func(context.Context, int) (models.Order, error) {
			return models.Order{}, service.ErrNotFound
		},
	}
	req := orderRequestWithID(http.MethodGet, "/orders/404", "404", "")
	w := httptest.NewRecorder()

	New(orderTestLogger(), mock).GetOrderByID(w, req)

	assertOrderStatus(t, w, http.StatusNotFound)
}

func TestGetOrdersByUserEmail(t *testing.T) {
	mock := &ordersMock{
		getOrdersByUserEmailFunc: func(_ context.Context, email string) ([]models.Order, error) {
			if email != "Alex@Example.com" {
				t.Fatalf("unexpected email: %s", email)
			}
			return []models.Order{}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/users/orders?email=Alex@Example.com", nil)
	w := httptest.NewRecorder()

	New(orderTestLogger(), mock).GetOrdersByUserEmail(w, req)

	assertOrderStatus(t, w, http.StatusOK)
}

func TestGetOrdersByUserEmailFailure(t *testing.T) {
	mock := &ordersMock{
		getOrdersByUserEmailFunc: func(context.Context, string) ([]models.Order, error) {
			return nil, errors.New("database unavailable")
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/users/orders?email=alex@example.com", nil)
	w := httptest.NewRecorder()

	New(orderTestLogger(), mock).GetOrdersByUserEmail(w, req)

	assertOrderStatus(t, w, http.StatusInternalServerError)
}

func orderRequestWithID(method, target, id, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func orderTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func assertOrderStatus(t *testing.T, recorder *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if recorder.Code != expected {
		t.Fatalf("expected status %d, got %d; body: %s", expected, recorder.Code, recorder.Body.String())
	}
}
