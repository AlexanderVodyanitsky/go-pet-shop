package product

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

//go:generate go run github.com/vektra/mockery/v2 --name=Products
type Products interface {
	GetAllProducts(ctx context.Context) ([]models.Product, error)
	GetProductByID(ctx context.Context, id int) (models.Product, error)
	CreateProduct(ctx context.Context, product models.Product) (models.Product, error)
	DeleteProduct(ctx context.Context, id int) error
	UpdateProduct(ctx context.Context, id int, product models.Product) (models.Product, error)
}

type Handler struct {
	log     *slog.Logger
	service Products
}

func New(log *slog.Logger, service Products) *Handler {
	return &Handler{log: log, service: service}
}

func (h *Handler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.product.GetAllProducts")

	products, err := h.service.GetAllProducts(r.Context())
	if err != nil {
		log.Error("failed to get products", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to retrieve products")
		return
	}

	httpx.JSON(w, http.StatusOK, products)
}

func (h *Handler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.product.GetProductByID")

	id, err := productID(r)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.service.GetProductByID(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		log.Error("failed to get product", slog.Int("id", id), slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to retrieve product")
		return
	}

	httpx.JSON(w, http.StatusOK, product)
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.product.CreateProduct")

	var product models.Product
	if err := httpx.DecodeJSON(r, &product); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	createdProduct, err := h.service.CreateProduct(r.Context(), product)
	if errors.Is(err, service.ErrInvalidInput) {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		log.Error("failed to create product", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to create product")
		return
	}

	httpx.JSON(w, http.StatusCreated, createdProduct)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.product.UpdateProduct")

	id, err := productID(r)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	var product models.Product
	if err := httpx.DecodeJSON(r, &product); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	updatedProduct, err := h.service.UpdateProduct(r.Context(), id, product)
	if errors.Is(err, service.ErrInvalidInput) {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		log.Error("failed to update product", slog.Int("id", id), slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to update product")
		return
	}

	httpx.JSON(w, http.StatusOK, updatedProduct)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.product.DeleteProduct")

	id, err := productID(r)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.DeleteProduct(r.Context(), id)
	if errors.Is(err, service.ErrInvalidInput) {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		log.Error("failed to delete product", slog.Int("id", id), slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) requestLogger(r *http.Request, fn string) *slog.Logger {
	return h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
}

func productID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		return 0, errors.New("product ID must be a positive integer")
	}
	return id, nil
}
