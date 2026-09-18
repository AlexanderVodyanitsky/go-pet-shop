package product

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	serviceerrors "go-pet-shop/internal/service"
	"testing"
)

type repositoryStub struct {
	createProductFunc func(context.Context, models.Product) (int, error)
}

func (r *repositoryStub) GetAllProducts(context.Context) ([]models.Product, error) {
	return nil, nil
}

func (r *repositoryStub) GetProductByID(context.Context, int) (models.Product, error) {
	return models.Product{}, nil
}

func (r *repositoryStub) CreateProduct(ctx context.Context, product models.Product) (int, error) {
	return r.createProductFunc(ctx, product)
}

func (r *repositoryStub) DeleteProduct(context.Context, int) error {
	return nil
}

func (r *repositoryStub) UpdateProduct(context.Context, models.Product) error {
	return nil
}

func TestCreateProductNormalizesAndValidates(t *testing.T) {
	repository := &repositoryStub{
		createProductFunc: func(_ context.Context, product models.Product) (int, error) {
			if product.Name != "Cat Food" {
				t.Fatalf("unexpected normalized name: %q", product.Name)
			}
			return 12, nil
		},
	}

	product, err := New(repository).CreateProduct(context.Background(), models.Product{
		Name: " Cat Food ", Price: 10, Stock: 2,
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}
	if product.ID != 12 || product.Name != "Cat Food" {
		t.Fatalf("unexpected product: %+v", product)
	}
}

func TestCreateProductRejectsInvalidDataBeforeRepository(t *testing.T) {
	called := false
	repository := &repositoryStub{
		createProductFunc: func(context.Context, models.Product) (int, error) {
			called = true
			return 0, nil
		},
	}

	_, err := New(repository).CreateProduct(context.Background(), models.Product{Name: " "})
	if !errors.Is(err, serviceerrors.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if called {
		t.Fatal("repository must not be called for invalid product")
	}
}
