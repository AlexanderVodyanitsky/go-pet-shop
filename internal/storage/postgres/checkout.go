package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"sort"

	"github.com/jackc/pgx/v5"
)

// PlaceOrder оформляет заказ атомарно: либо остатки, заказ, позиции и платёж
// сохраняются вместе, либо транзакция откатывает все изменения.
func (s *Storage) PlaceOrder(
	ctx context.Context,
	userEmail string,
	items []models.OrderItem,
) (int, error) {
	const fn = "storage.postgres.checkout.PlaceOrder"

	if userEmail == "" || len(items) == 0 {
		return 0, fmt.Errorf("%s: %w", fn, storage.ErrInvalidInput)
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("%s: begin transaction: %w", fn, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var userExists bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`,
		userEmail,
	).Scan(&userExists); err != nil {
		return 0, fmt.Errorf("%s: check user: %w", fn, err)
	}
	if !userExists {
		return 0, fmt.Errorf("%s: %w: user=%s", fn, storage.ErrNotFound, userEmail)
	}

	normalizedItems, err := combineOrderItems(items)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	var totalPrice float64
	for _, item := range normalizedItems {
		var price float64
		err := tx.QueryRow(ctx, `
			UPDATE products
			SET stock = stock - $1
			WHERE id = $2 AND stock >= $1
			RETURNING price
		`, item.Quantity, item.ProductID).Scan(&price)
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, classifyUnavailableProduct(ctx, tx, fn, item.ProductID)
		}
		if err != nil {
			return 0, fmt.Errorf("%s: update product %d stock: %w", fn, item.ProductID, err)
		}
		totalPrice += price * float64(item.Quantity)
	}

	var orderID int
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (user_email, total_price)
		VALUES ($1, $2)
		RETURNING id
	`, userEmail, totalPrice).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("%s: create order: %w", fn, err)
	}

	for _, item := range normalizedItems {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity)
			VALUES ($1, $2, $3)
		`, orderID, item.ProductID, item.Quantity)
		if err != nil {
			return 0, fmt.Errorf("%s: add order item: %w", fn, err)
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (order_id, amount, status)
		VALUES ($1, $2, 'completed')
	`, orderID, totalPrice)
	if err != nil {
		return 0, fmt.Errorf("%s: create payment transaction: %w", fn, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("%s: commit transaction: %w", fn, err)
	}

	return orderID, nil
}

func combineOrderItems(items []models.OrderItem) ([]models.OrderItem, error) {
	quantities := make(map[int]int, len(items))
	order := make([]int, 0, len(items))
	for _, item := range items {
		if item.ProductID < 1 || item.Quantity < 1 {
			return nil, storage.ErrInvalidInput
		}
		if _, exists := quantities[item.ProductID]; !exists {
			order = append(order, item.ProductID)
		}
		quantities[item.ProductID] += item.Quantity
		if quantities[item.ProductID] < 1 {
			return nil, storage.ErrInvalidInput
		}
	}
	sort.Ints(order)

	result := make([]models.OrderItem, 0, len(order))
	for _, productID := range order {
		result = append(result, models.OrderItem{
			ProductID: productID,
			Quantity:  quantities[productID],
		})
	}
	return result, nil
}

func classifyUnavailableProduct(
	ctx context.Context,
	tx pgx.Tx,
	fn string,
	productID int,
) error {
	var exists bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)`,
		productID,
	).Scan(&exists); err != nil {
		return fmt.Errorf("%s: check product %d: %w", fn, productID, err)
	}
	if !exists {
		return fmt.Errorf("%s: %w: product=%d", fn, storage.ErrNotFound, productID)
	}
	return fmt.Errorf("%s: %w: product=%d", fn, storage.ErrInsufficientStock, productID)
}
