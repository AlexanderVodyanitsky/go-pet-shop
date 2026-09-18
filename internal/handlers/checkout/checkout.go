package checkout

import (
	"context"
	"errors"
	"go-pet-shop/internal/handlers/httpx"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/service"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
)

type Orders interface {
	PlaceOrder(ctx context.Context, userEmail string, items []models.OrderItem) (int, error)
}

type Handler struct {
	log     *slog.Logger
	service Orders
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

func New(log *slog.Logger, service Orders) *Handler {
	return &Handler{log: log, service: service}
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
	items := make([]models.OrderItem, 0, len(checkoutRequest.Items))
	for _, item := range checkoutRequest.Items {
		items = append(items, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	orderID, err := h.service.PlaceOrder(r.Context(), checkoutRequest.UserEmail, items)
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		httpx.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrNotFound):
		httpx.Error(w, http.StatusNotFound, "user or product not found")
	case errors.Is(err, service.ErrInsufficientStock):
		httpx.Error(w, http.StatusConflict, "insufficient product stock")
	case err != nil:
		log.Error("failed to place order", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to place order")
	default:
		httpx.JSON(w, http.StatusCreated, response{OrderID: orderID, Status: "completed"})
	}
}
