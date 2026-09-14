package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// CreateOrder создаёт пустой заказ и возвращает назначенный базой ID.
func (s *Storage) CreateOrder(ctx context.Context, order models.Order) (int, error) {
	const fn = "storage.postgres.order.CreateOrder"

	var id int
	err := s.db.QueryRow(ctx,
		`INSERT INTO orders (user_email, total_price) VALUES ($1, $2) RETURNING id`,
		order.UserEmail, order.TotalPrice,
	).Scan(&id)
	if err != nil {
		return 0, orderWriteError(fn, err)
	}

	return id, nil
}

// AddOrderItem добавляет товар в существующий заказ.
func (s *Storage) AddOrderItem(ctx context.Context, item models.OrderItem) error {
	const fn = "storage.postgres.order.AddOrderItem"

	_, err := s.db.Exec(ctx,
		`INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)`,
		item.OrderID, item.ProductID, item.Quantity,
	)
	if err != nil {
		return orderWriteError(fn, err)
	}

	return nil
}

// GetOrderByID получает основные данные заказа.
func (s *Storage) GetOrderByID(ctx context.Context, id int) (models.Order, error) {
	const fn = "storage.postgres.order.GetOrderByID"

	var order models.Order
	err := s.db.QueryRow(ctx, `
		SELECT id, user_email, total_price, created_at
		FROM orders
		WHERE id = $1
	`, id).Scan(&order.ID, &order.UserEmail, &order.TotalPrice, &order.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Order{}, fmt.Errorf("%s: %w: id=%d", fn, storage.ErrNotFound, id)
	}
	if err != nil {
		return models.Order{}, fmt.Errorf("%s: %w", fn, err)
	}

	return order, nil
}

// GetOrdersByUserEmail получает все заказы пользователя. JOIN гарантирует, что
// возвращаются только заказы существующего пользователя.
func (s *Storage) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error) {
	const fn = "storage.postgres.order.GetOrdersByUserEmail"

	rows, err := s.db.Query(ctx, `
		SELECT o.id, o.user_email, o.total_price, o.created_at
		FROM orders AS o
		JOIN users AS u ON u.email = o.user_email
		WHERE u.email = $1
		ORDER BY o.created_at DESC, o.id DESC
	`, email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
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

// GetOrderItemsByOrderID получает позиции заказа вместе с данными товара.
func (s *Storage) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	const fn = "storage.postgres.order.GetOrderItemsByOrderID"

	rows, err := s.db.Query(ctx, `
		SELECT oi.id, oi.order_id, oi.product_id, p.name, p.price, oi.quantity
		FROM order_items AS oi
		JOIN products AS p ON p.id = oi.product_id
		WHERE oi.order_id = $1
		ORDER BY oi.id
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	items := make([]models.OrderItem, 0)
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.ProductPrice,
			&item.Quantity,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return items, nil
}

func orderWriteError(fn string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		case "23514":
			return fmt.Errorf("%s: %w", fn, storage.ErrInvalidInput)
		}
	}
	return fmt.Errorf("%s: %w", fn, err)
}
