package order

import (
	"context"
	"errors"
	"go-pet-shop/internal/handlers/httpx"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/service"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Orders interface {
	CreateOrder(ctx context.Context, order models.Order) (models.Order, error)
	AddOrderItem(ctx context.Context, orderID int, item models.OrderItem) (models.OrderItem, error)
	GetOrderByID(ctx context.Context, id int) (models.Order, error)
	GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error)
}

type Handler struct {
	log     *slog.Logger
	service Orders
}

type createOrderRequest struct {
	UserEmail  string  `json:"user_email"`
	TotalPrice float64 `json:"total_price"`
}

type addOrderItemRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type createOrderResponse struct {
	ID         int     `json:"id"`
	UserEmail  string  `json:"user_email"`
	TotalPrice float64 `json:"total_price"`
}

func New(log *slog.Logger, service Orders) *Handler {
	return &Handler{log: log, service: service}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.order.CreateOrder")

	var request createOrderRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	createdOrder, err := h.service.CreateOrder(r.Context(), models.Order{
		UserEmail:  request.UserEmail,
		TotalPrice: request.TotalPrice,
	})
	if errors.Is(err, service.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if errors.Is(err, service.ErrInvalidInput) {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		log.Error("failed to create order", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	httpx.JSON(w, http.StatusCreated, createOrderResponse{
		ID:         createdOrder.ID,
		UserEmail:  createdOrder.UserEmail,
		TotalPrice: createdOrder.TotalPrice,
	})
}

func (h *Handler) AddOrderItem(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.order.AddOrderItem")

	orderID, err := orderID(r)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	var request addOrderItemRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	item, err := h.service.AddOrderItem(r.Context(), orderID, models.OrderItem{
		ProductID: request.ProductID,
		Quantity:  request.Quantity,
	})
	if errors.Is(err, service.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "order or product not found")
		return
	}
	if errors.Is(err, service.ErrInvalidInput) {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		log.Error("failed to add order item", slog.Int("order_id", orderID), slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to add order item")
		return
	}

	httpx.JSON(w, http.StatusCreated, item)
}

func (h *Handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.order.GetOrderByID")

	id, err := orderID(r)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.service.GetOrderByID(r.Context(), id)
	if errors.Is(err, service.ErrInvalidInput) {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		log.Error("failed to get order", slog.Int("order_id", id), slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to retrieve order")
		return
	}

	httpx.JSON(w, http.StatusOK, order)
}

func (h *Handler) GetOrdersByUserEmail(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.order.GetOrdersByUserEmail")

	email := chi.URLParam(r, "email")
	if email == "" {
		email = r.URL.Query().Get("email")
	}
	orders, err := h.service.GetOrdersByUserEmail(r.Context(), email)
	if errors.Is(err, service.ErrInvalidInput) {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		log.Error("failed to get user orders", slog.String("email", email), slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to retrieve user orders")
		return
	}

	httpx.JSON(w, http.StatusOK, orders)
}

func (h *Handler) requestLogger(r *http.Request, fn string) *slog.Logger {
	return h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
}

func orderID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		return 0, errors.New("order ID must be a positive integer")
	}
	return id, nil
}
