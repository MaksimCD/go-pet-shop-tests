package shop

import (
	"context"
	"encoding/json"
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

type Store interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
	CreateProduct(ctx context.Context, product models.Product) (int, error)
	GetProductByID(ctx context.Context, id int) (models.Product, error)
	GetAllProducts(ctx context.Context) ([]models.Product, error)
	UpdateProduct(ctx context.Context, product models.Product) error
	DeleteProduct(ctx context.Context, id int) error
	CreateOrder(ctx context.Context, order models.Order) (int, error)
	AddOrderItem(ctx context.Context, orderItem models.OrderItem) error
	GetOrderByID(ctx context.Context, id int) (models.Order, error)
	GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error)
	PlaceOrder(ctx context.Context, userEmail string, items []models.OrderItem) (int, error)
	GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error)
	GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error)
}

type Handler struct {
	log   *slog.Logger
	store Store
}

type checkoutRequest struct {
	UserEmail string             `json:"user_email"`
	Items     []models.OrderItem `json:"items"`
}

func New(log *slog.Logger, store Store) *Handler {
	return &Handler{log: log, store: store}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := decodeJSON(r, &user); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(user.Name) == "" || strings.TrimSpace(user.Email) == "" {
		writeError(w, r, http.StatusBadRequest, "name and email are required")
		return
	}

	id, err := h.store.CreateUser(r.Context(), user)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	user.ID = id
	writeJSON(w, r, http.StatusCreated, user)
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.GetAllUsers(r.Context())
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, users)
}

func (h *Handler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	user, err := h.store.GetUserByEmail(r.Context(), email)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, user)
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := decodeJSON(r, &product); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(product.Name) == "" {
		writeError(w, r, http.StatusBadRequest, "product name is required")
		return
	}
	if product.Price < 0 || product.Stock < 0 {
		writeError(w, r, http.StatusBadRequest, "price and stock must be non-negative")
		return
	}

	id, err := h.store.CreateProduct(r.Context(), product)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	product.ID = id
	writeJSON(w, r, http.StatusCreated, product)
}

func (h *Handler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.store.GetAllProducts(r.Context())
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, products)
}

func (h *Handler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	product, err := h.store.GetProductByID(r.Context(), id)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, product)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	var product models.Product
	if err := decodeJSON(r, &product); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(product.Name) == "" {
		writeError(w, r, http.StatusBadRequest, "product name is required")
		return
	}
	if product.Price < 0 || product.Stock < 0 {
		writeError(w, r, http.StatusBadRequest, "price and stock must be non-negative")
		return
	}

	product.ID = id
	if err := h.store.UpdateProduct(r.Context(), product); err != nil {
		writeStoreError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, product)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	if err := h.store.DeleteProduct(r.Context(), id); err != nil {
		writeStoreError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := decodeJSON(r, &order); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(order.UserEmail) == "" {
		writeError(w, r, http.StatusBadRequest, "user_email is required")
		return
	}

	id, err := h.store.CreateOrder(r.Context(), order)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	order.ID = id
	writeJSON(w, r, http.StatusCreated, order)
}

func (h *Handler) AddOrderItem(w http.ResponseWriter, r *http.Request) {
	orderID, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	var item models.OrderItem
	if err := decodeJSON(r, &item); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if item.ProductID <= 0 || item.Quantity <= 0 {
		writeError(w, r, http.StatusBadRequest, "product_id and quantity must be positive")
		return
	}

	item.OrderID = orderID
	if err := h.store.AddOrderItem(r.Context(), item); err != nil {
		writeStoreError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusCreated, item)
}

func (h *Handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	order, err := h.store.GetOrderByID(r.Context(), id)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	items, err := h.store.GetOrderItemsByOrderID(r.Context(), id)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, map[string]any{
		"order": order,
		"items": items,
	})
}

func (h *Handler) GetOrdersByUserEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	if email == "" {
		email = r.URL.Query().Get("email")
	}
	if strings.TrimSpace(email) == "" {
		writeError(w, r, http.StatusBadRequest, "email is required")
		return
	}

	orders, err := h.store.GetOrdersByUserEmail(r.Context(), email)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, orders)
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req checkoutRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.UserEmail) == "" || len(req.Items) == 0 {
		writeError(w, r, http.StatusBadRequest, "user_email and items are required")
		return
	}

	orderID, err := h.store.PlaceOrder(r.Context(), req.UserEmail, req.Items)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusCreated, map[string]int{"order_id": orderID})
}

func (h *Handler) GetUserOrderHistory(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	if email == "" {
		email = r.URL.Query().Get("email")
	}
	if strings.TrimSpace(email) == "" {
		writeError(w, r, http.StatusBadRequest, "email is required")
		return
	}

	history, err := h.store.GetUserOrderHistory(r.Context(), email)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, history)
}

func (h *Handler) GetPopularProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.store.GetPopularProducts(r.Context())
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, products)
}

func parseIDParam(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, name))
	if err != nil || id <= 0 {
		writeError(w, r, http.StatusBadRequest, name+" must be a positive integer")
		return 0, false
	}

	return id, true
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return errors.New("invalid JSON payload")
	}

	return nil
}

func writeStoreError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, storage.ErrNotFound):
		writeError(w, r, http.StatusNotFound, "resource not found")
	case errors.Is(err, storage.ErrConflict):
		writeError(w, r, http.StatusConflict, "resource already exists")
	case errors.Is(err, storage.ErrInvalidInput):
		writeError(w, r, http.StatusBadRequest, "invalid input")
	case errors.Is(err, storage.ErrInsufficientStock):
		writeError(w, r, http.StatusConflict, "insufficient stock")
	default:
		writeError(w, r, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	writeJSON(w, r, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
