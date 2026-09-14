package product

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func requestWithID(method, target, id, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", id)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	return req, httptest.NewRecorder()
}

func request(method, target, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	return req, httptest.NewRecorder()
}

func assertStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()

	if recorder.Code != want {
		t.Fatalf("expected status %d, got %d", want, recorder.Code)
	}
}

// =======================
// Get All Products
// =======================

func TestGetAllProducts_Success(t *testing.T) {
	mock := &ProductsMock{
		GetAllProductsFunc: func(ctx context.Context) ([]models.Product, error) {
			return []models.Product{
				{ID: 1, Name: "Dog Food", Price: 10.5, Stock: 3},
			}, nil
		},
	}

	req, w := request(http.MethodGet, "/products", "")
	handler := New(testLogger(), mock)

	handler.GetAllProducts(w, req)

	assertStatus(t, w, http.StatusOK)
}

func TestGetAllProducts_Fail(t *testing.T) {
	mock := &ProductsMock{
		GetAllProductsFunc: func(ctx context.Context) ([]models.Product, error) {
			return nil, errors.New("db error")
		},
	}

	req, w := request(http.MethodGet, "/products", "")
	handler := New(testLogger(), mock)

	handler.GetAllProducts(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}

func TestGetProductByID_Success(t *testing.T) {
	mock := &ProductsMock{
		GetProductByIDFunc: func(ctx context.Context, id int) (models.Product, error) {
			if id != 1 {
				t.Fatalf("expected id 1, got %d", id)
			}
			return models.Product{ID: 1, Name: "Dog Food", Price: 10.5, Stock: 3}, nil
		},
	}

	req, w := requestWithID(http.MethodGet, "/products/1", "1", "")
	New(testLogger(), mock).GetProductByID(w, req)

	assertStatus(t, w, http.StatusOK)
}

func TestGetProductByID_NotFound(t *testing.T) {
	mock := &ProductsMock{
		GetProductByIDFunc: func(ctx context.Context, id int) (models.Product, error) {
			return models.Product{}, storage.ErrNotFound
		},
	}

	req, w := requestWithID(http.MethodGet, "/products/42", "42", "")
	New(testLogger(), mock).GetProductByID(w, req)

	assertStatus(t, w, http.StatusNotFound)
}

// =======================
// Create Product
// =======================

func TestCreateProduct_Success(t *testing.T) {
	called := false
	mock := &ProductsMock{
		CreateProductFunc: func(ctx context.Context, product models.Product) (int, error) {
			called = true

			if product.Name != "Dog Food" || product.Price != 10.5 || product.Stock != 3 {
				t.Fatalf("unexpected product: %+v", product)
			}

			return 1, nil
		},
	}

	req, w := request(http.MethodPost, "/products", `{"Name":"Dog Food","Price":10.5,"Stock":3}`)
	handler := New(testLogger(), mock)

	handler.CreateProduct(w, req)

	assertStatus(t, w, http.StatusCreated)
	if !called {
		t.Fatal("expected CreateProduct to be called")
	}
}

func TestCreateProduct_BadRequest(t *testing.T) {
	mock := &ProductsMock{
		CreateProductFunc: func(ctx context.Context, product models.Product) (int, error) {
			t.Fatal("CreateProduct must not be called for invalid JSON")
			return 0, nil
		},
	}

	req, w := request(http.MethodPost, "/products", `{"Name":`)
	handler := New(testLogger(), mock)

	handler.CreateProduct(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}

func TestCreateProduct_Fail(t *testing.T) {
	mock := &ProductsMock{
		CreateProductFunc: func(ctx context.Context, product models.Product) (int, error) {
			return 0, errors.New("db error")
		},
	}

	req, w := request(http.MethodPost, "/products", `{"Name":"Dog Food","Price":10.5,"Stock":3}`)
	handler := New(testLogger(), mock)

	handler.CreateProduct(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}

// =======================
// Update Product
// =======================

func TestUpdateProduct_Success(t *testing.T) {
	called := false
	mock := &ProductsMock{
		UpdateProductFunc: func(ctx context.Context, product models.Product) error {
			called = true

			if product.ID != 1 || product.Name != "Cat Toy" || product.Price != 7.25 || product.Stock != 8 {
				t.Fatalf("unexpected product: %+v", product)
			}

			return nil
		},
	}

	req, w := requestWithID(http.MethodPut, "/products/1", "1", `{"Name":"Cat Toy","Price":7.25,"Stock":8}`)
	handler := New(testLogger(), mock)

	handler.UpdateProduct(w, req)

	assertStatus(t, w, http.StatusOK)
	if !called {
		t.Fatal("expected UpdateProduct to be called")
	}
}

func TestUpdateProduct_BadRequest(t *testing.T) {
	mock := &ProductsMock{
		UpdateProductFunc: func(ctx context.Context, product models.Product) error {
			t.Fatal("UpdateProduct must not be called for invalid JSON")
			return nil
		},
	}

	req, w := requestWithID(http.MethodPut, "/products/1", "1", `{"Name":`)
	handler := New(testLogger(), mock)

	handler.UpdateProduct(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}

func TestUpdateProduct_Fail(t *testing.T) {
	mock := &ProductsMock{
		UpdateProductFunc: func(ctx context.Context, product models.Product) error {
			return errors.New("db error")
		},
	}

	req, w := requestWithID(http.MethodPut, "/products/1", "1", `{"Name":"Cat Toy","Price":7.25,"Stock":8}`)
	handler := New(testLogger(), mock)

	handler.UpdateProduct(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}

// =======================
// Delete Product
// =======================

func TestDeleteProduct_Success(t *testing.T) {
	called := false
	mock := &ProductsMock{
		DeleteProductFunc: func(ctx context.Context, id int) error {
			called = true

			if id != 1 {
				t.Fatalf("expected id 1, got %d", id)
			}

			return nil
		},
	}

	req, w := requestWithID(http.MethodDelete, "/products/1", "1", "")
	handler := New(testLogger(), mock)

	handler.DeleteProduct(w, req)

	assertStatus(t, w, http.StatusNoContent)
	if !called {
		t.Fatal("expected DeleteProduct to be called")
	}
}

func TestDeleteProduct_BadRequest(t *testing.T) {
	mock := &ProductsMock{
		DeleteProductFunc: func(ctx context.Context, id int) error {
			t.Fatal("DeleteProduct must not be called for empty id")
			return nil
		},
	}

	req, w := request(http.MethodDelete, "/products/", "")
	handler := New(testLogger(), mock)

	handler.DeleteProduct(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}

func TestDeleteProduct_Fail(t *testing.T) {
	mock := &ProductsMock{
		DeleteProductFunc: func(ctx context.Context, id int) error {
			return errors.New("db error")
		},
	}

	req, w := requestWithID(http.MethodDelete, "/products/1", "1", "")
	handler := New(testLogger(), mock)

	handler.DeleteProduct(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}
