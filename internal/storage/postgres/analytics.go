package postgres

import (
	"context"
	"fmt"
	"go-pet-shop/internal/models"
)

// GetUserOrderHistory возвращает заказы, позиции и состояние оплаты пользователя.
func (s *Storage) GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error) {
	const fn = "storage.postgres.analytics.GetUserOrderHistory"

	rows, err := s.db.Query(ctx, `
		SELECT
			o.id,
			o.user_email,
			o.total_price,
			o.created_at,
			COALESCE(t.amount, 0),
			COALESCE(t.status, 'not_recorded'),
			COALESCE(oi.id, 0),
			COALESCE(oi.product_id, 0),
			COALESCE(p.name, ''),
			COALESCE(p.price, 0),
			COALESCE(oi.quantity, 0)
		FROM orders AS o
		LEFT JOIN transactions AS t ON t.order_id = o.id
		LEFT JOIN order_items AS oi ON oi.order_id = o.id
		LEFT JOIN products AS p ON p.id = oi.product_id
		WHERE o.user_email = $1
		ORDER BY o.created_at DESC, o.id DESC, oi.id
	`, email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	history := make([]models.OrderDetail, 0)
	orderIndexes := make(map[int]int)
	for rows.Next() {
		var detail models.OrderDetail
		var item models.OrderItem
		if err := rows.Scan(
			&detail.OrderID,
			&detail.UserEmail,
			&detail.TotalPrice,
			&detail.CreatedAt,
			&detail.TransactionAmount,
			&detail.TransactionStatus,
			&item.ID,
			&item.ProductID,
			&item.ProductName,
			&item.ProductPrice,
			&item.Quantity,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		index, exists := orderIndexes[detail.OrderID]
		if !exists {
			detail.Items = make([]models.OrderItem, 0)
			history = append(history, detail)
			index = len(history) - 1
			orderIndexes[detail.OrderID] = index
		}
		if item.ID > 0 {
			item.OrderID = detail.OrderID
			history[index].Items = append(history[index].Items, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return history, nil
}

// GetPopularProducts ранжирует проданные товары по суммарному количеству.
func (s *Storage) GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error) {
	const fn = "storage.postgres.analytics.GetPopularProducts"

	rows, err := s.db.Query(ctx, `
		SELECT p.id, p.name, p.price, p.stock, SUM(oi.quantity)::BIGINT AS total_quantity
		FROM products AS p
		JOIN order_items AS oi ON oi.product_id = p.id
		JOIN transactions AS t ON t.order_id = oi.order_id AND t.status = 'completed'
		GROUP BY p.id, p.name, p.price, p.stock
		ORDER BY total_quantity DESC, p.id
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	products := make([]models.PopularProduct, 0)
	for rows.Next() {
		var product models.PopularProduct
		if err := rows.Scan(
			&product.ProductID,
			&product.Name,
			&product.Price,
			&product.Stock,
			&product.TotalQuantity,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return products, nil
}
