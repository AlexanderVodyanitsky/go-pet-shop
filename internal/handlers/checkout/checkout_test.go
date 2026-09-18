package checkout

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
)

type checkoutMock struct {
	placeOrderFunc func(context.Context, string, []models.OrderItem) (int, error)
}

func (m *checkoutMock) PlaceOrder(
	ctx context.Context,
	email string,
	items []models.OrderItem,
) (int, error) {
	return m.placeOrderFunc(ctx, email, items)
}

func TestPlaceOrder(t *testing.T) {
	mock := &checkoutMock{
		placeOrderFunc: func(_ context.Context, email string, items []models.OrderItem) (int, error) {
			if email != "Alex@Example.com" {
				t.Fatalf("unexpected email: %s", email)
			}
			if len(items) != 1 || items[0].ProductID != 3 || items[0].Quantity != 2 {
				t.Fatalf("unexpected items: %+v", items)
			}
			return 17, nil
		},
	}
	req := checkoutRequest(`{
		"user_email":"Alex@Example.com",
		"items":[{"product_id":3,"quantity":2}]
	}`)
	w := httptest.NewRecorder()

	New(checkoutTestLogger(), mock).PlaceOrder(w, req)

	assertCheckoutStatus(t, w, http.StatusCreated)
	if !strings.Contains(w.Body.String(), `"order_id":17`) {
		t.Fatalf("unexpected response: %s", w.Body.String())
	}
}

func TestPlaceOrderRejectsEmptyItems(t *testing.T) {
	mock := &checkoutMock{
		placeOrderFunc: func(context.Context, string, []models.OrderItem) (int, error) {
			return 0, service.InvalidInput("at least one order item is required")
		},
	}
	req := checkoutRequest(`{"user_email":"alex@example.com","items":[]}`)
	w := httptest.NewRecorder()

	New(checkoutTestLogger(), mock).PlaceOrder(w, req)

	assertCheckoutStatus(t, w, http.StatusBadRequest)
}

func TestPlaceOrderNotFound(t *testing.T) {
	mock := checkoutErrorMock(service.ErrNotFound)
	w := httptest.NewRecorder()

	New(checkoutTestLogger(), mock).PlaceOrder(w, validCheckoutRequest())

	assertCheckoutStatus(t, w, http.StatusNotFound)
}

func TestPlaceOrderInsufficientStock(t *testing.T) {
	mock := checkoutErrorMock(service.ErrInsufficientStock)
	w := httptest.NewRecorder()

	New(checkoutTestLogger(), mock).PlaceOrder(w, validCheckoutRequest())

	assertCheckoutStatus(t, w, http.StatusConflict)
}

func TestPlaceOrderInternalFailure(t *testing.T) {
	mock := checkoutErrorMock(errors.New("commit failed"))
	w := httptest.NewRecorder()

	New(checkoutTestLogger(), mock).PlaceOrder(w, validCheckoutRequest())

	assertCheckoutStatus(t, w, http.StatusInternalServerError)
}

func checkoutErrorMock(err error) *checkoutMock {
	return &checkoutMock{
		placeOrderFunc: func(context.Context, string, []models.OrderItem) (int, error) {
			return 0, err
		},
	}
}

func validCheckoutRequest() *http.Request {
	return checkoutRequest(`{
		"user_email":"alex@example.com",
		"items":[{"product_id":3,"quantity":2}]
	}`)
}

func checkoutRequest(body string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/checkout", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func checkoutTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func assertCheckoutStatus(t *testing.T, recorder *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if recorder.Code != expected {
		t.Fatalf("expected status %d, got %d; body: %s", expected, recorder.Code, recorder.Body.String())
	}
}
