package checkout

import (
	"context"
	"errors"
	"go-pet-shop/internal/handlers/httpx"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"github.com/go-chi/chi/middleware"
)

type Orders interface {
	PlaceOrder(ctx context.Context, userEmail string, items []models.OrderItem) (int, error)
}

type Handler struct {
	log     *slog.Logger
	storage Orders
}

type request struct {
	UserEmail string        `json:"user_email"`
	Items     []itemRequest `json:"items"`
}

type itemRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type response struct {
	OrderID int    `json:"order_id"`
	Status  string `json:"status"`
}

func New(log *slog.Logger, storage Orders) *Handler {
	return &Handler{log: log, storage: storage}
}

func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	log := h.log.With(
		slog.String("fn", "handlers.checkout.PlaceOrder"),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var checkoutRequest request
	if err := httpx.DecodeJSON(r, &checkoutRequest); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	checkoutRequest.UserEmail = strings.ToLower(strings.TrimSpace(checkoutRequest.UserEmail))
	if !validEmail(checkoutRequest.UserEmail) {
		httpx.Error(w, http.StatusBadRequest, "valid user email is required")
		return
	}
	if len(checkoutRequest.Items) == 0 {
		httpx.Error(w, http.StatusBadRequest, "at least one order item is required")
		return
	}

	items := make([]models.OrderItem, 0, len(checkoutRequest.Items))
	for _, item := range checkoutRequest.Items {
		if item.ProductID < 1 {
			httpx.Error(w, http.StatusBadRequest, "product ID must be a positive integer")
			return
		}
		if item.Quantity < 1 {
			httpx.Error(w, http.StatusBadRequest, "quantity must be a positive integer")
			return
		}
		items = append(items, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	orderID, err := h.storage.PlaceOrder(r.Context(), checkoutRequest.UserEmail, items)
	switch {
	case errors.Is(err, storage.ErrInvalidInput):
		httpx.Error(w, http.StatusBadRequest, "invalid checkout data")
	case errors.Is(err, storage.ErrNotFound):
		httpx.Error(w, http.StatusNotFound, "user or product not found")
	case errors.Is(err, storage.ErrInsufficientStock):
		httpx.Error(w, http.StatusConflict, "insufficient product stock")
	case err != nil:
		log.Error("failed to place order", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to place order")
	default:
		httpx.JSON(w, http.StatusCreated, response{OrderID: orderID, Status: "completed"})
	}
}

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}
