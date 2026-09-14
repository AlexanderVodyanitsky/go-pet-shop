package order

import (
	"context"
	"errors"
	"go-pet-shop/internal/handlers/httpx"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Orders interface {
	CreateOrder(ctx context.Context, order models.Order) (int, error)
	AddOrderItem(ctx context.Context, item models.OrderItem) error
	GetOrderByID(ctx context.Context, id int) (models.Order, error)
	GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error)
}

type Handler struct {
	log     *slog.Logger
	storage Orders
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

func New(log *slog.Logger, storage Orders) *Handler {
	return &Handler{log: log, storage: storage}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.order.CreateOrder")

	var request createOrderRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	request.UserEmail = strings.ToLower(strings.TrimSpace(request.UserEmail))
	if !validEmail(request.UserEmail) {
		httpx.Error(w, http.StatusBadRequest, "valid user email is required")
		return
	}
	if request.TotalPrice < 0 {
		httpx.Error(w, http.StatusBadRequest, "order total price cannot be negative")
		return
	}

	order := models.Order{UserEmail: request.UserEmail, TotalPrice: request.TotalPrice}
	id, err := h.storage.CreateOrder(r.Context(), order)
	if errors.Is(err, storage.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if errors.Is(err, storage.ErrInvalidInput) {
		httpx.Error(w, http.StatusBadRequest, "invalid order data")
		return
	}
	if err != nil {
		log.Error("failed to create order", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	httpx.JSON(w, http.StatusCreated, createOrderResponse{
		ID:         id,
		UserEmail:  order.UserEmail,
		TotalPrice: order.TotalPrice,
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
	if request.ProductID < 1 {
		httpx.Error(w, http.StatusBadRequest, "product ID must be a positive integer")
		return
	}
	if request.Quantity < 1 {
		httpx.Error(w, http.StatusBadRequest, "quantity must be a positive integer")
		return
	}

	item := models.OrderItem{
		OrderID:   orderID,
		ProductID: request.ProductID,
		Quantity:  request.Quantity,
	}
	err = h.storage.AddOrderItem(r.Context(), item)
	if errors.Is(err, storage.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "order or product not found")
		return
	}
	if errors.Is(err, storage.ErrInvalidInput) {
		httpx.Error(w, http.StatusBadRequest, "invalid order item")
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

	order, err := h.storage.GetOrderByID(r.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		log.Error("failed to get order", slog.Int("order_id", id), slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to retrieve order")
		return
	}

	items, err := h.storage.GetOrderItemsByOrderID(r.Context(), id)
	if err != nil {
		log.Error("failed to get order items", slog.Int("order_id", id), slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to retrieve order items")
		return
	}
	order.Items = items

	httpx.JSON(w, http.StatusOK, order)
}

func (h *Handler) GetOrdersByUserEmail(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.order.GetOrdersByUserEmail")

	email := chi.URLParam(r, "email")
	if email == "" {
		email = r.URL.Query().Get("email")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if !validEmail(email) {
		httpx.Error(w, http.StatusBadRequest, "valid user email is required")
		return
	}

	orders, err := h.storage.GetOrdersByUserEmail(r.Context(), email)
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

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}
