package user

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

	"github.com/go-chi/chi"
)

type usersMock struct {
	createUserFunc     func(context.Context, models.User) (models.User, error)
	getUserByEmailFunc func(context.Context, string) (models.User, error)
	getAllUsersFunc    func(context.Context) ([]models.User, error)
}

func (m *usersMock) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	return m.createUserFunc(ctx, user)
}

func (m *usersMock) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	return m.getUserByEmailFunc(ctx, email)
}

func (m *usersMock) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return m.getAllUsersFunc(ctx)
}

func TestCreateUser(t *testing.T) {
	mock := &usersMock{
		createUserFunc: func(_ context.Context, user models.User) (models.User, error) {
			if user.Name != " Alex " || user.Email != "Alex@Example.com" {
				t.Fatalf("unexpected user: %+v", user)
			}
			return models.User{ID: 7, Name: "Alex", Email: "alex@example.com"}, nil
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(
		`{"name":" Alex ","email":"Alex@Example.com"}`,
	))
	w := httptest.NewRecorder()

	New(discardLogger(), mock).CreateUser(w, req)

	assertResponseStatus(t, w, http.StatusCreated)
}

func TestCreateUserRejectsInvalidEmail(t *testing.T) {
	mock := &usersMock{
		createUserFunc: func(context.Context, models.User) (models.User, error) {
			return models.User{}, service.InvalidInput("valid user email is required")
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(
		`{"name":"Alex","email":"not-an-email"}`,
	))
	w := httptest.NewRecorder()

	New(discardLogger(), mock).CreateUser(w, req)

	assertResponseStatus(t, w, http.StatusBadRequest)
}

func TestCreateUserConflict(t *testing.T) {
	mock := &usersMock{
		createUserFunc: func(context.Context, models.User) (models.User, error) {
			return models.User{}, service.ErrConflict
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(
		`{"name":"Alex","email":"alex@example.com"}`,
	))
	w := httptest.NewRecorder()

	New(discardLogger(), mock).CreateUser(w, req)

	assertResponseStatus(t, w, http.StatusConflict)
}

func TestGetAllUsers(t *testing.T) {
	mock := &usersMock{
		getAllUsersFunc: func(context.Context) ([]models.User, error) {
			return []models.User{{ID: 1, Name: "Alex", Email: "alex@example.com"}}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	New(discardLogger(), mock).GetAllUsers(w, req)

	assertResponseStatus(t, w, http.StatusOK)
}

func TestGetUserByEmailNotFound(t *testing.T) {
	mock := &usersMock{
		getUserByEmailFunc: func(context.Context, string) (models.User, error) {
			return models.User{}, service.ErrNotFound
		},
	}
	req := requestWithEmail("missing@example.com")
	w := httptest.NewRecorder()

	New(discardLogger(), mock).GetUserByEmail(w, req)

	assertResponseStatus(t, w, http.StatusNotFound)
}

func TestGetAllUsersFailure(t *testing.T) {
	mock := &usersMock{
		getAllUsersFunc: func(context.Context) ([]models.User, error) {
			return nil, errors.New("database unavailable")
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	New(discardLogger(), mock).GetAllUsers(w, req)

	assertResponseStatus(t, w, http.StatusInternalServerError)
}

func requestWithEmail(email string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/users/"+email, nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("email", email)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func assertResponseStatus(t *testing.T, recorder *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if recorder.Code != expected {
		t.Fatalf("expected status %d, got %d; body: %s", expected, recorder.Code, recorder.Body.String())
	}
}
