package user

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	serviceerrors "go-pet-shop/internal/service"
	"testing"
)

type repositoryStub struct {
	createUserFunc func(context.Context, models.User) (int, error)
}

func (r *repositoryStub) CreateUser(ctx context.Context, user models.User) (int, error) {
	return r.createUserFunc(ctx, user)
}

func (r *repositoryStub) GetUserByEmail(context.Context, string) (models.User, error) {
	return models.User{}, nil
}

func (r *repositoryStub) GetAllUsers(context.Context) ([]models.User, error) {
	return nil, nil
}

func TestCreateUserNormalizesDomainData(t *testing.T) {
	repository := &repositoryStub{
		createUserFunc: func(_ context.Context, user models.User) (int, error) {
			if user.Name != "Alex" || user.Email != "alex@example.com" {
				t.Fatalf("unexpected normalized user: %+v", user)
			}
			return 5, nil
		},
	}

	user, err := New(repository).CreateUser(context.Background(), models.User{
		Name: " Alex ", Email: "Alex@Example.com",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if user.ID != 5 {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestCreateUserRejectsInvalidEmailBeforeRepository(t *testing.T) {
	called := false
	repository := &repositoryStub{
		createUserFunc: func(context.Context, models.User) (int, error) {
			called = true
			return 0, nil
		},
	}

	_, err := New(repository).CreateUser(context.Background(), models.User{
		Name: "Alex", Email: "invalid",
	})
	if !errors.Is(err, serviceerrors.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if called {
		t.Fatal("repository must not be called for invalid user")
	}
}
