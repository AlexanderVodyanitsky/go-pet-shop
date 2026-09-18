package user

import (
	"context"
	"go-pet-shop/internal/models"
	serviceerrors "go-pet-shop/internal/service"
	"strings"
)

type Repository interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	user.Name = strings.TrimSpace(user.Name)
	if user.Name == "" {
		return models.User{}, serviceerrors.InvalidInput("user name is required")
	}

	email, err := serviceerrors.NormalizeEmail(user.Email)
	if err != nil {
		return models.User{}, err
	}
	user.Email = email

	id, err := s.repository.CreateUser(ctx, user)
	if err != nil {
		return models.User{}, serviceerrors.MapStorageError(err)
	}
	user.ID = id

	return user, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	normalizedEmail, err := serviceerrors.NormalizeEmail(email)
	if err != nil {
		return models.User{}, err
	}

	user, err := s.repository.GetUserByEmail(ctx, normalizedEmail)
	return user, serviceerrors.MapStorageError(err)
}

func (s *Service) GetAllUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.repository.GetAllUsers(ctx)
	return users, serviceerrors.MapStorageError(err)
}
