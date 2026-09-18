package user

import (
	"context"
	"errors"
	"go-pet-shop/internal/handlers/httpx"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/service"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Users interface {
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type Handler struct {
	log     *slog.Logger
	service Users
}

func New(log *slog.Logger, service Users) *Handler {
	return &Handler{log: log, service: service}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.user.CreateUser")

	var user models.User
	if err := httpx.DecodeJSON(r, &user); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	createdUser, err := h.service.CreateUser(r.Context(), user)
	if errors.Is(err, service.ErrInvalidInput) {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrConflict) {
		httpx.Error(w, http.StatusConflict, "a user with this email already exists")
		return
	}
	if err != nil {
		log.Error("failed to create user", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	httpx.JSON(w, http.StatusCreated, createdUser)
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.user.GetAllUsers")

	users, err := h.service.GetAllUsers(r.Context())
	if err != nil {
		log.Error("failed to get users", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to retrieve users")
		return
	}

	httpx.JSON(w, http.StatusOK, users)
}

func (h *Handler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.user.GetUserByEmail")

	email := chi.URLParam(r, "email")
	value, err := h.service.GetUserByEmail(r.Context(), email)
	if errors.Is(err, service.ErrInvalidInput) {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Error("failed to get user", slog.String("email", email), slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to retrieve user")
		return
	}

	httpx.JSON(w, http.StatusOK, value)
}

func (h *Handler) requestLogger(r *http.Request, fn string) *slog.Logger {
	return h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
}
