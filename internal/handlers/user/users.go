package user

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

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Users interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type Handler struct {
	log     *slog.Logger
	storage Users
}

func New(log *slog.Logger, storage Users) *Handler {
	return &Handler{log: log, storage: storage}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.user.CreateUser")

	var user models.User
	if err := httpx.DecodeJSON(r, &user); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	if err := validateUser(&user); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.storage.CreateUser(r.Context(), user)
	if errors.Is(err, storage.ErrConflict) {
		httpx.Error(w, http.StatusConflict, "a user with this email already exists")
		return
	}
	if err != nil {
		log.Error("failed to create user", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	user.ID = id
	httpx.JSON(w, http.StatusCreated, user)
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.user.GetAllUsers")

	users, err := h.storage.GetAllUsers(r.Context())
	if err != nil {
		log.Error("failed to get users", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, "failed to retrieve users")
		return
	}

	httpx.JSON(w, http.StatusOK, users)
}

func (h *Handler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	log := h.requestLogger(r, "handlers.user.GetUserByEmail")

	email := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "email")))
	if !validEmail(email) {
		httpx.Error(w, http.StatusBadRequest, "valid user email is required")
		return
	}

	value, err := h.storage.GetUserByEmail(r.Context(), email)
	if errors.Is(err, storage.ErrNotFound) {
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

func validateUser(user *models.User) error {
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	if user.Name == "" {
		return errors.New("user name is required")
	}
	if !validEmail(user.Email) {
		return errors.New("valid user email is required")
	}
	return nil
}

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}
