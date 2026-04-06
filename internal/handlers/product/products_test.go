package product

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	testifymock "github.com/stretchr/testify/mock"
)

func addURLParam(req *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func TestGetAllProducts_Success(t *testing.T) {
	mockSvc := NewProductsMock(t)
	mockSvc.On("GetAllProducts", testifymock.Anything).Return([]models.Product{
		{ID: 1, Name: "Dog Food", Price: 10.5, Stock: 5},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mockSvc)
	handler.GetAllProducts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestGetAllProducts_Error(t *testing.T) {
	mockSvc := NewProductsMock(t)
	mockSvc.On("GetAllProducts", testifymock.Anything).Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mockSvc)
	handler.GetAllProducts(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestCreateProduct_Success(t *testing.T) {
	mockSvc := NewProductsMock(t)
	mockSvc.On("CreateProduct", testifymock.Anything, models.Product{
		Name:  "Cat Food",
		Price: 15.5,
		Stock: 7,
	}).Return(10, nil)

	body := `{"name":"Cat Food","price":15.5,"stock":7}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mockSvc)
	handler.CreateProduct(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestCreateProduct_BadRequest(t *testing.T) {
	mockSvc := NewProductsMock(t)

	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(`{"name":`))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mockSvc)
	handler.CreateProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCreateProduct_Fail(t *testing.T) {
	mockSvc := NewProductsMock(t)
	mockSvc.On("CreateProduct", testifymock.Anything, models.Product{
		Name:  "Cat Food",
		Price: 15.5,
		Stock: 7,
	}).Return(0, errors.New("service error"))

	body := `{"name":"Cat Food","price":15.5,"stock":7}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mockSvc)
	handler.CreateProduct(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestUpdateProduct_Success(t *testing.T) {
	mockSvc := NewProductsMock(t)
	mockSvc.On("UpdateProduct", testifymock.Anything, models.Product{
		ID:    5,
		Name:  "Updated Food",
		Price: 20.0,
		Stock: 3,
	}).Return(nil)

	body := `{"name":"Updated Food","price":20.0,"stock":3}`
	req := httptest.NewRequest(http.MethodPut, "/products/5", strings.NewReader(body))
	req = addURLParam(req, "id", "5")
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mockSvc)
	handler.UpdateProduct(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUpdateProduct_BadRequest(t *testing.T) {
	mockSvc := NewProductsMock(t)

	req := httptest.NewRequest(http.MethodPut, "/products/5", strings.NewReader(`{"name":`))
	req = addURLParam(req, "id", "5")
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mockSvc)
	handler.UpdateProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateProduct_Fail(t *testing.T) {
	mockSvc := NewProductsMock(t)
	mockSvc.On("UpdateProduct", testifymock.Anything, models.Product{
		ID:    5,
		Name:  "Updated Food",
		Price: 20.0,
		Stock: 3,
	}).Return(errors.New("service error"))

	body := `{"name":"Updated Food","price":20.0,"stock":3}`
	req := httptest.NewRequest(http.MethodPut, "/products/5", strings.NewReader(body))
	req = addURLParam(req, "id", "5")
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mockSvc)
	handler.UpdateProduct(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestDeleteProduct_Success(t *testing.T) {
	mockSvc := NewProductsMock(t)
	mockSvc.On("DeleteProduct", testifymock.Anything, 7).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/products/7", nil)
	req = addURLParam(req, "id", "7")
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mockSvc)
	handler.DeleteProduct(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestDeleteProduct_BadRequest(t *testing.T) {
	mockSvc := NewProductsMock(t)

	req := httptest.NewRequest(http.MethodDelete, "/products/", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mockSvc)
	handler.DeleteProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteProduct_Fail(t *testing.T) {
	mockSvc := NewProductsMock(t)
	mockSvc.On("DeleteProduct", testifymock.Anything, 7).Return(errors.New("service error"))

	req := httptest.NewRequest(http.MethodDelete, "/products/7", nil)
	req = addURLParam(req, "id", "7")
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mockSvc)
	handler.DeleteProduct(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
