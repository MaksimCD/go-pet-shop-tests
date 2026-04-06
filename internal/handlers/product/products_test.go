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
)

func addURLParam(req *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

// Get Product - Ready
func TestGetAllProducts_Success(t *testing.T) {
	// Мокаем storage — он вернёт один продукт.
	mock := &ProductsMock{
		GetAllProductsFunc: func(ctx context.Context) ([]models.Product, error) {
			return []models.Product{
				{ID: 1, Name: "Dog Food", Price: 10.5, Stock: 5},
			}, nil
		},
	}

	// Создаем HTTP-запрос GET /products
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	// Создаем хендлер с мок-хранилищем
	handler := New(slog.Default(), mock)

	// Вызываем метод GetAllProducts, который является http.HandlerFunc
	handler.GetAllProducts(w, req)

	// Проверяем HTTP-код
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
func TestGetAllProducts_Error(t *testing.T) {
	// Мокаем storage — он будет возвращать ошибку
	mock := &ProductsMock{
		GetAllProductsFunc: func(ctx context.Context) ([]models.Product, error) {
			return nil, errors.New("db error")
		},
	}

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)
	handler.GetAllProducts(w, req)

	// Ожидаем HTTP 500
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// =======================
// Create Product
// =======================

func TestCreateProduct_Success(t *testing.T) {
	mock := &ProductsMock{
		CreateProductFunc: func(ctx context.Context, product models.Product) (int, error) {
			if product.Name != "Cat Food" {
				t.Fatalf("expected product name Cat Food, got %s", product.Name)
			}
			return 10, nil
		},
	}
	body := `{"name":"Cat Food", "price":15.5,"stock":7}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	w := httptest.NewRecorder()

	heandler := New(slog.Default(), mock)
	heandler.CreateProduct(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestCreateProduct_BadRequest(t *testing.T) {
	mock := &ProductsMock{}

	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(`{"name":`))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)
	handler.CreateProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf(" expected status 400, got %d", w.Code)
	}
}

func TestCreateProduct_Fail(t *testing.T) {
	mock := &ProductsMock{
		CreateProductFunc: func(ctx context.Context, product models.Product) (int, error) {
			return 0, errors.New("service error")
		},
	}

	body := `{"name":"Cat Food","price":15.5,"stock":7}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)
	handler.CreateProduct(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

// =======================
// Update Product
// =======================

func TestUpdateProduct_Success(t *testing.T) {
	mock := &ProductsMock{
		UpdateProductFunc: func(ctx context.Context, product models.Product) error {
			if product.ID != 5 {
				t.Fatalf("expeted product name Updated Food, got %s", product.Name)
			}
			return nil
		},
	}

	body := `{"name":"Updated Food", "price":20.0,"stock":3}`
	req := httptest.NewRequest(http.MethodPut, "/products/5", strings.NewReader(body))
	req = addURLParam(req, "id", "5")
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)
	handler.UpdateProduct(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUpdateProduct_BadRequest(t *testing.T) {
	mock := &ProductsMock{}

	req := httptest.NewRequest(http.MethodPut, "/products/5", strings.NewReader(`{"name":`))
	req = addURLParam(req, "id", "5")
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)
	handler.UpdateProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateProduct_Fail(t *testing.T) {
	mock := &ProductsMock{
		UpdateProductFunc: func(ctx context.Context, product models.Product) error {
			return errors.New("service error")
		},
	}

	body := `{"name":"Updated Food", "price":20.0, "stock":3}`
	req := httptest.NewRequest(http.MethodPut, "/products/5", strings.NewReader(body))
	req = addURLParam(req, "id", "5")
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)
	handler.UpdateProduct(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

// =======================
// Delete Product
// =======================

func TestDeleteProduct_Success(t *testing.T) {
	mock := &ProductsMock{
		DeleteProductFunc: func(ctx context.Context, id int) error {
			if id != 7 {
				t.Fatalf("expected id 7, got %d", id)
			}
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/products/7", nil)
	req = addURLParam(req, "id", "7")
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)
	handler.DeleteProduct(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestDeleteProduct_BadRequest(t *testing.T) {
	mock := &ProductsMock{}

	req := httptest.NewRequest(http.MethodDelete, "/products/", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)
	handler.DeleteProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteProduct_Fail(t *testing.T) {
	mock := &ProductsMock{
		DeleteProductFunc: func(ctx context.Context, id int) error {
			return errors.New("service error")
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/products/7", nil)
	req = addURLParam(req, "id", "7")
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)
	handler.DeleteProduct(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
