package product

import (
	"context"
	"errors"
	productmocks "go-pet-shop/internal/handlers/product/mocks"
	"go-pet-shop/internal/models"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/mock"
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
	storageMock := productmocks.NewProducts(t)
	storageMock.On("GetAllProducts", mock.Anything).Return([]models.Product{
		{ID: 1, Name: "Dog Food", Price: 10.5, Stock: 3},
	}, nil).Once()

	req, w := request(http.MethodGet, "/products", "")
	handler := New(testLogger(), storageMock)

	handler.GetAllProducts(w, req)

	assertStatus(t, w, http.StatusOK)
}

func TestGetAllProducts_Fail(t *testing.T) {
	storageMock := productmocks.NewProducts(t)
	storageMock.On("GetAllProducts", mock.Anything).Return(nil, errors.New("db error")).Once()

	req, w := request(http.MethodGet, "/products", "")
	handler := New(testLogger(), storageMock)

	handler.GetAllProducts(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}

// =======================
// Create Product
// =======================

func TestCreateProduct_Success(t *testing.T) {
	storageMock := productmocks.NewProducts(t)
	storageMock.On("CreateProduct", mock.Anything, models.Product{
		Name:  "Dog Food",
		Price: 10.5,
		Stock: 3,
	}).Return(1, nil).Once()

	req, w := request(http.MethodPost, "/products", `{"Name":"Dog Food","Price":10.5,"Stock":3}`)
	handler := New(testLogger(), storageMock)

	handler.CreateProduct(w, req)

	assertStatus(t, w, http.StatusOK)
}

func TestCreateProduct_BadRequest(t *testing.T) {
	storageMock := productmocks.NewProducts(t)

	req, w := request(http.MethodPost, "/products", `{"Name":`)
	handler := New(testLogger(), storageMock)

	handler.CreateProduct(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}

func TestCreateProduct_Fail(t *testing.T) {
	storageMock := productmocks.NewProducts(t)
	storageMock.On("CreateProduct", mock.Anything, models.Product{
		Name:  "Dog Food",
		Price: 10.5,
		Stock: 3,
	}).Return(0, errors.New("db error")).Once()

	req, w := request(http.MethodPost, "/products", `{"Name":"Dog Food","Price":10.5,"Stock":3}`)
	handler := New(testLogger(), storageMock)

	handler.CreateProduct(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}

// =======================
// Update Product
// =======================

func TestUpdateProduct_Success(t *testing.T) {
	storageMock := productmocks.NewProducts(t)
	storageMock.On("UpdateProduct", mock.Anything, models.Product{
		ID:    1,
		Name:  "Cat Toy",
		Price: 7.25,
		Stock: 8,
	}).Return(nil).Once()

	req, w := requestWithID(http.MethodPut, "/products/1", "1", `{"Name":"Cat Toy","Price":7.25,"Stock":8}`)
	handler := New(testLogger(), storageMock)

	handler.UpdateProduct(w, req)

	assertStatus(t, w, http.StatusOK)
}

func TestUpdateProduct_BadRequest(t *testing.T) {
	storageMock := productmocks.NewProducts(t)

	req, w := requestWithID(http.MethodPut, "/products/1", "1", `{"Name":`)
	handler := New(testLogger(), storageMock)

	handler.UpdateProduct(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}

func TestUpdateProduct_Fail(t *testing.T) {
	storageMock := productmocks.NewProducts(t)
	storageMock.On("UpdateProduct", mock.Anything, models.Product{
		ID:    1,
		Name:  "Cat Toy",
		Price: 7.25,
		Stock: 8,
	}).Return(errors.New("db error")).Once()

	req, w := requestWithID(http.MethodPut, "/products/1", "1", `{"Name":"Cat Toy","Price":7.25,"Stock":8}`)
	handler := New(testLogger(), storageMock)

	handler.UpdateProduct(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}

// =======================
// Delete Product
// =======================

func TestDeleteProduct_Success(t *testing.T) {
	storageMock := productmocks.NewProducts(t)
	storageMock.On("DeleteProduct", mock.Anything, 1).Return(nil).Once()

	req, w := requestWithID(http.MethodDelete, "/products/1", "1", "")
	handler := New(testLogger(), storageMock)

	handler.DeleteProduct(w, req)

	assertStatus(t, w, http.StatusOK)
}

func TestDeleteProduct_BadRequest(t *testing.T) {
	storageMock := productmocks.NewProducts(t)

	req, w := request(http.MethodDelete, "/products/", "")
	handler := New(testLogger(), storageMock)

	handler.DeleteProduct(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}

func TestDeleteProduct_Fail(t *testing.T) {
	storageMock := productmocks.NewProducts(t)
	storageMock.On("DeleteProduct", mock.Anything, 1).Return(errors.New("db error")).Once()

	req, w := requestWithID(http.MethodDelete, "/products/1", "1", "")
	handler := New(testLogger(), storageMock)

	handler.DeleteProduct(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}
