package analytics

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
)

type analyticsMock struct {
	getHistoryFunc func(context.Context, string) ([]models.OrderDetail, error)
	getPopularFunc func(context.Context) ([]models.PopularProduct, error)
}

func (m *analyticsMock) GetUserOrderHistory(
	ctx context.Context,
	email string,
) ([]models.OrderDetail, error) {
	if m.getHistoryFunc == nil {
		return nil, nil
	}
	return m.getHistoryFunc(ctx, email)
}

func (m *analyticsMock) GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error) {
	if m.getPopularFunc == nil {
		return nil, nil
	}
	return m.getPopularFunc(ctx)
}

func TestGetUserOrderHistoryFromQuery(t *testing.T) {
	mock := &analyticsMock{
		getHistoryFunc: func(_ context.Context, email string) ([]models.OrderDetail, error) {
			if email != "alex@example.com" {
				t.Fatalf("unexpected email: %s", email)
			}
			return []models.OrderDetail{{
				OrderID:           1,
				UserEmail:         email,
				TotalPrice:        20,
				CreatedAt:         time.Now(),
				TransactionAmount: 20,
				TransactionStatus: "completed",
				Items:             []models.OrderItem{},
			}}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/users/history?email=Alex@Example.com", nil)
	w := httptest.NewRecorder()

	New(analyticsTestLogger(), mock).GetUserOrderHistory(w, req)

	assertAnalyticsStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"transaction_status":"completed"`) {
		t.Fatalf("unexpected response: %s", w.Body.String())
	}
}

func TestGetUserOrderHistoryFromPath(t *testing.T) {
	mock := &analyticsMock{
		getHistoryFunc: func(_ context.Context, email string) ([]models.OrderDetail, error) {
			if email != "alex@example.com" {
				t.Fatalf("unexpected email: %s", email)
			}
			return []models.OrderDetail{}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/users/alex@example.com/history", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("email", "alex@example.com")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	w := httptest.NewRecorder()

	New(analyticsTestLogger(), mock).GetUserOrderHistory(w, req)

	assertAnalyticsStatus(t, w, http.StatusOK)
}

func TestGetUserOrderHistoryRejectsInvalidEmail(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/history?email=invalid", nil)
	w := httptest.NewRecorder()

	New(analyticsTestLogger(), &analyticsMock{}).GetUserOrderHistory(w, req)

	assertAnalyticsStatus(t, w, http.StatusBadRequest)
}

func TestGetPopularProducts(t *testing.T) {
	mock := &analyticsMock{
		getPopularFunc: func(context.Context) ([]models.PopularProduct, error) {
			return []models.PopularProduct{{
				ProductID: 1, Name: "Cat Food", Price: 10, Stock: 4, TotalQuantity: 8,
			}}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/products/popular", nil)
	w := httptest.NewRecorder()

	New(analyticsTestLogger(), mock).GetPopularProducts(w, req)

	assertAnalyticsStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"total_quantity":8`) {
		t.Fatalf("unexpected response: %s", w.Body.String())
	}
}

func TestGetPopularProductsFailure(t *testing.T) {
	mock := &analyticsMock{
		getPopularFunc: func(context.Context) ([]models.PopularProduct, error) {
			return nil, errors.New("database unavailable")
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/products/popular", nil)
	w := httptest.NewRecorder()

	New(analyticsTestLogger(), mock).GetPopularProducts(w, req)

	assertAnalyticsStatus(t, w, http.StatusInternalServerError)
}

func analyticsTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func assertAnalyticsStatus(t *testing.T, recorder *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if recorder.Code != expected {
		t.Fatalf("expected status %d, got %d; body: %s", expected, recorder.Code, recorder.Body.String())
	}
}
