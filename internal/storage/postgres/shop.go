package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateUser(ctx context.Context, user models.User) (int, error) {
	const fn = "storage.postgres.CreateUser"

	var id int
	err := s.db.QueryRow(
		ctx,
		`INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`,
		user.Name,
		user.Email,
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrConflict)
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	const fn = "storage.postgres.GetUserByEmail"

	var user models.User
	err := s.db.QueryRow(
		ctx,
		`SELECT id, name, email FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", fn, err)
	}

	return user, nil
}

func (s *Storage) GetAllUsers(ctx context.Context) ([]models.User, error) {
	const fn = "storage.postgres.GetAllUsers"

	rows, err := s.db.Query(ctx, `SELECT id, name, email FROM users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return users, nil
}

func (s *Storage) CreateProduct(ctx context.Context, product models.Product) (int, error) {
	const fn = "storage.postgres.CreateProduct"

	var id int
	err := s.db.QueryRow(
		ctx,
		`INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id`,
		product.Name,
		product.Price,
		product.Stock,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetProductByID(ctx context.Context, id int) (models.Product, error) {
	const fn = "storage.postgres.GetProductByID"

	var product models.Product
	err := s.db.QueryRow(
		ctx,
		`SELECT id, name, price, stock FROM products WHERE id = $1`,
		id,
	).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Product{}, fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		}
		return models.Product{}, fmt.Errorf("%s: %w", fn, err)
	}

	return product, nil
}

func (s *Storage) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	const fn = "storage.postgres.GetAllProducts"

	rows, err := s.db.Query(ctx, `SELECT id, name, price, stock FROM products ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return products, nil
}

func (s *Storage) UpdateProduct(ctx context.Context, product models.Product) error {
	const fn = "storage.postgres.UpdateProduct"

	cmd, err := s.db.Exec(
		ctx,
		`UPDATE products SET name = $1, price = $2, stock = $3 WHERE id = $4`,
		product.Name,
		product.Price,
		product.Stock,
		product.ID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
	}

	return nil
}

func (s *Storage) DeleteProduct(ctx context.Context, id int) error {
	const fn = "storage.postgres.DeleteProduct"

	cmd, err := s.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
	}

	return nil
}

func (s *Storage) CreateOrder(ctx context.Context, order models.Order) (int, error) {
	const fn = "storage.postgres.CreateOrder"

	var id int
	err := s.db.QueryRow(
		ctx,
		`INSERT INTO orders (user_email, total_price) VALUES ($1, $2) RETURNING id`,
		order.UserEmail,
		order.TotalPrice,
	).Scan(&id)
	if err != nil {
		if isForeignKeyViolation(err) {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) AddOrderItem(ctx context.Context, orderItem models.OrderItem) error {
	const fn = "storage.postgres.AddOrderItem"

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer tx.Rollback(ctx)

	var price float64
	err = tx.QueryRow(ctx, `SELECT price FROM products WHERE id = $1`, orderItem.ProductID).Scan(&price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)`,
		orderItem.OrderID,
		orderItem.ProductID,
		orderItem.Quantity,
	); err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	if _, err := tx.Exec(
		ctx,
		`UPDATE orders SET total_price = total_price + $1 WHERE id = $2`,
		price*float64(orderItem.Quantity),
		orderItem.OrderID,
	); err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Storage) GetOrderByID(ctx context.Context, id int) (models.Order, error) {
	const fn = "storage.postgres.GetOrderByID"

	var order models.Order
	err := s.db.QueryRow(
		ctx,
		`SELECT id, user_email, total_price, created_at FROM orders WHERE id = $1`,
		id,
	).Scan(&order.ID, &order.UserEmail, &order.TotalPrice, &order.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		}
		return models.Order{}, fmt.Errorf("%s: %w", fn, err)
	}

	return order, nil
}

func (s *Storage) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error) {
	const fn = "storage.postgres.GetOrdersByUserEmail"

	rows, err := s.db.Query(
		ctx,
		`SELECT id, user_email, total_price, created_at
		 FROM orders
		 WHERE user_email = $1
		 ORDER BY created_at DESC, id DESC`,
		email,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.UserEmail, &order.TotalPrice, &order.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return orders, nil
}

func (s *Storage) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	const fn = "storage.postgres.GetOrderItemsByOrderID"

	rows, err := s.db.Query(
		ctx,
		`SELECT id, order_id, product_id, quantity
		 FROM order_items
		 WHERE order_id = $1
		 ORDER BY id`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return items, nil
}

func (s *Storage) PlaceOrder(ctx context.Context, userEmail string, items []models.OrderItem) (int, error) {
	const fn = "storage.postgres.PlaceOrder"

	if len(items) == 0 {
		return 0, fmt.Errorf("%s: %w", fn, storage.ErrInvalidInput)
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	defer tx.Rollback(ctx)

	if err := ensureUserExists(ctx, tx, userEmail); err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	totalPrice := 0.0
	for _, item := range items {
		if item.Quantity <= 0 {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrInvalidInput)
		}

		var price float64
		err := tx.QueryRow(
			ctx,
			`UPDATE products
			 SET stock = stock - $2
			 WHERE id = $1 AND stock >= $2
			 RETURNING price`,
			item.ProductID,
			item.Quantity,
		).Scan(&price)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				exists, enough, checkErr := productAvailability(ctx, tx, item.ProductID, item.Quantity)
				if checkErr != nil {
					return 0, fmt.Errorf("%s: %w", fn, checkErr)
				}
				if !exists {
					return 0, fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
				}
				if !enough {
					return 0, fmt.Errorf("%s: %w", fn, storage.ErrInsufficientStock)
				}
			}
			return 0, fmt.Errorf("%s: %w", fn, err)
		}

		totalPrice += price * float64(item.Quantity)
	}

	var orderID int
	err = tx.QueryRow(
		ctx,
		`INSERT INTO orders (user_email, total_price) VALUES ($1, $2) RETURNING id`,
		userEmail,
		totalPrice,
	).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	for _, item := range items {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)`,
			orderID,
			item.ProductID,
			item.Quantity,
		); err != nil {
			return 0, fmt.Errorf("%s: %w", fn, err)
		}
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO transactions (order_id, amount, status) VALUES ($1, $2, $3)`,
		orderID,
		totalPrice,
		"paid",
	); err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return orderID, nil
}

func (s *Storage) GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error) {
	const fn = "storage.postgres.GetUserOrderHistory"

	rows, err := s.db.Query(
		ctx,
		`SELECT
			o.id,
			o.user_email,
			o.total_price,
			o.created_at,
			oi.id,
			oi.order_id,
			oi.product_id,
			oi.quantity,
			COALESCE(p.name, ''),
			COALESCE(p.price, 0),
			t.id,
			t.order_id,
			t.amount,
			t.status,
			t.created_at
		 FROM orders o
		 LEFT JOIN order_items oi ON oi.order_id = o.id
		 LEFT JOIN products p ON p.id = oi.product_id
		 LEFT JOIN transactions t ON t.order_id = o.id
		 WHERE o.user_email = $1
		 ORDER BY o.created_at DESC, o.id DESC, oi.id ASC`,
		email,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	var history []models.OrderDetail
	indexByOrderID := make(map[int]int)

	for rows.Next() {
		var (
			order       models.Order
			itemID      sql.NullInt64
			itemOrderID sql.NullInt64
			productID   sql.NullInt64
			quantity    sql.NullInt64
			productName string
			unitPrice   float64
			txID        sql.NullInt64
			txOrderID   sql.NullInt64
			txAmount    sql.NullFloat64
			txStatus    sql.NullString
			txCreatedAt sql.NullTime
		)

		if err := rows.Scan(
			&order.ID,
			&order.UserEmail,
			&order.TotalPrice,
			&order.CreatedAt,
			&itemID,
			&itemOrderID,
			&productID,
			&quantity,
			&productName,
			&unitPrice,
			&txID,
			&txOrderID,
			&txAmount,
			&txStatus,
			&txCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		idx, ok := indexByOrderID[order.ID]
		if !ok {
			history = append(history, models.OrderDetail{Order: order})
			idx = len(history) - 1
			indexByOrderID[order.ID] = idx
		}

		if itemID.Valid && itemOrderID.Valid && productID.Valid && quantity.Valid {
			history[idx].Items = append(history[idx].Items, models.OrderItemDetail{
				ID:          int(itemID.Int64),
				OrderID:     int(itemOrderID.Int64),
				ProductID:   int(productID.Int64),
				ProductName: productName,
				Quantity:    int(quantity.Int64),
				UnitPrice:   unitPrice,
				LineTotal:   unitPrice * float64(quantity.Int64),
			})
		}

		if txID.Valid && txOrderID.Valid && txAmount.Valid && txStatus.Valid && txCreatedAt.Valid {
			history[idx].Transaction = &models.Transaction{
				ID:        int(txID.Int64),
				OrderID:   int(txOrderID.Int64),
				Amount:    txAmount.Float64,
				Status:    txStatus.String,
				CreatedAt: txCreatedAt.Time,
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return history, nil
}

func (s *Storage) GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error) {
	const fn = "storage.postgres.GetPopularProducts"

	rows, err := s.db.Query(
		ctx,
		`SELECT
			p.id,
			p.name,
			COALESCE(SUM(oi.quantity), 0) AS total_sold
		 FROM products p
		 LEFT JOIN order_items oi ON oi.product_id = p.id
		 GROUP BY p.id, p.name
		 ORDER BY total_sold DESC, p.id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	var products []models.PopularProduct
	for rows.Next() {
		var product models.PopularProduct
		if err := rows.Scan(&product.ProductID, &product.Name, &product.TotalSold); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return products, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func ensureUserExists(ctx context.Context, tx pgx.Tx, email string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return storage.ErrNotFound
	}

	return nil
}

func productAvailability(ctx context.Context, tx pgx.Tx, productID, quantity int) (bool, bool, error) {
	var stock int
	err := tx.QueryRow(ctx, `SELECT stock FROM products WHERE id = $1`, productID).Scan(&stock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, false, nil
		}
		return false, false, err
	}

	return true, stock >= quantity, nil
}
