package analytics

import (
	"context"
	"go-pet-shop/internal/handlers/httpx"
	"go-pet-shop/internal/models"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Analytics interface {
	GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error)
	GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error)
}

type Handler struct {
	log     *slog.Logger
	storage Analytics
}

func New(log *slog.Logger, storage Analytics) *Handler {
	return &Handler{log: log, storage: storage}
}

func (h *Handler) GetUserOrderHistory(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.analytics.GetUserOrderHistory")

	email := chi.URLParam(r, "email")
	if email == "" {
		email = r.URL.Query().Get("email")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if !validEmail(email) {
		httpx.Error(w, http.StatusBadRequest, "valid user email is required")
		return
	}

	history, err := h.storage.GetUserOrderHistory(r.Context(), email)
	if err != nil {
		log.Error("failed to get user order history", slog.String("email", email), slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to retrieve order history")
		return
	}

	httpx.JSON(w, http.StatusOK, history)
}

func (h *Handler) GetPopularProducts(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.analytics.GetPopularProducts")

	products, err := h.storage.GetPopularProducts(r.Context())
	if err != nil {
		log.Error("failed to get popular products", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to retrieve popular products")
		return
	}

	httpx.JSON(w, http.StatusOK, products)
}

func (h *Handler) requestLogger(r *http.Request, fn string) *slog.Logger {
	return h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
}

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}
