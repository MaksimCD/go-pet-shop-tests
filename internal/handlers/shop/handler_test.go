package shop

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
)

type storeStub struct {
	createUserFn func(context.Context, models.User) (int, error)
	getProductFn func(context.Context, int) (models.Product, error)
	checkoutFn   func(context.Context, string, []models.OrderItem) (int, error)
}

func (s storeStub) CreateUser(ctx context.Context, user models.User) (int, error) {
	return s.createUserFn(ctx, user)
}
func (s storeStub) GetUserByEmail(context.Context, string) (models.User, error) {
	return models.User{}, nil
}
func (s storeStub) GetAllUsers(context.Context) ([]models.User, error) {
	return nil, nil
}
func (s storeStub) CreateProduct(context.Context, models.Product) (int, error) {
	return 0, nil
}
func (s storeStub) GetProductByID(ctx context.Context, id int) (models.Product, error) {
	return s.getProductFn(ctx, id)
}
func (s storeStub) GetAllProducts(context.Context) ([]models.Product, error) {
	return nil, nil
}
func (s storeStub) UpdateProduct(context.Context, models.Product) error {
	return nil
}
func (s storeStub) DeleteProduct(context.Context, int) error {
	return nil
}
func (s storeStub) CreateOrder(context.Context, models.Order) (int, error) {
	return 0, nil
}
func (s storeStub) AddOrderItem(context.Context, models.OrderItem) error {
	return nil
}
func (s storeStub) GetOrderByID(context.Context, int) (models.Order, error) {
	return models.Order{}, nil
}
func (s storeStub) GetOrdersByUserEmail(context.Context, string) ([]models.Order, error) {
	return nil, nil
}
func (s storeStub) GetOrderItemsByOrderID(context.Context, int) ([]models.OrderItem, error) {
	return nil, nil
}
func (s storeStub) PlaceOrder(ctx context.Context, email string, items []models.OrderItem) (int, error) {
	return s.checkoutFn(ctx, email, items)
}
func (s storeStub) GetUserOrderHistory(context.Context, string) ([]models.OrderDetail, error) {
	return nil, nil
}
func (s storeStub) GetPopularProducts(context.Context) ([]models.PopularProduct, error) {
	return nil, nil
}

func TestCreateUser(t *testing.T) {
	h := New(slog.Default(), storeStub{
		createUserFn: func(_ context.Context, user models.User) (int, error) {
			if user.Email != "max@example.com" {
				t.Fatalf("unexpected email: %s", user.Email)
			}
			return 7, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Max","email":"max@example.com"}`))
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestGetProductByIDNotFound(t *testing.T) {
	h := New(slog.Default(), storeStub{
		getProductFn: func(context.Context, int) (models.Product, error) {
			return models.Product{}, storage.ErrNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	req = withURLParam(req, "id", "1")
	rec := httptest.NewRecorder()

	h.GetProductByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCheckoutInsufficientStock(t *testing.T) {
	h := New(slog.Default(), storeStub{
		checkoutFn: func(context.Context, string, []models.OrderItem) (int, error) {
			return 0, storage.ErrInsufficientStock
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/checkout", strings.NewReader(`{"user_email":"max@example.com","items":[{"product_id":1,"quantity":3}]}`))
	rec := httptest.NewRecorder()

	h.Checkout(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestCheckoutBadJSON(t *testing.T) {
	h := New(slog.Default(), storeStub{
		checkoutFn: func(context.Context, string, []models.OrderItem) (int, error) {
			return 0, errors.New("should not be called")
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/checkout", strings.NewReader(`{"user_email":`))
	rec := httptest.NewRecorder()

	h.Checkout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func withURLParam(req *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}
