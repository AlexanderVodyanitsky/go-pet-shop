package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"

	"github.com/jackc/pgx/v5"
)

// GetAllProducts - получает все продукты
func (s *Storage) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	const fn = "storage.postgres.product.GetAllProducts"

	rows, err := s.db.Query(ctx, `SELECT id, name, price, stock FROM products ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		products = append(products, p)
	}

	// Проверяем ошибки итерации
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return products, nil
}

// GetProductByID получает товар по идентификатору.
func (s *Storage) GetProductByID(ctx context.Context, id int) (models.Product, error) {
	const fn = "storage.postgres.product.GetProductByID"

	var product models.Product
	err := s.db.QueryRow(ctx,
		`SELECT id, name, price, stock FROM products WHERE id = $1`,
		id,
	).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Product{}, fmt.Errorf("%s: %w: id=%d", fn, storage.ErrNotFound, id)
	}
	if err != nil {
		return models.Product{}, fmt.Errorf("%s: %w", fn, err)
	}

	return product, nil
}

// CreateProduct - создает продукт и возвращает его ID
func (s *Storage) CreateProduct(ctx context.Context, p models.Product) (int, error) {
	const fn = "storage.postgres.product.CreateProduct"

	var id int
	err := s.db.QueryRow(ctx,
		`INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id`,
		p.Name, p.Price, p.Stock).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

// DeleteProduct - удаляет продукт по ID
func (s *Storage) DeleteProduct(ctx context.Context, id int) error {
	const fn = "storage.postgres.product.DeleteProduct"

	cmd, err := s.db.Exec(ctx,
		`DELETE FROM products WHERE id = $1`,
		id)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w: id=%d", fn, storage.ErrNotFound, id)
	}

	return nil
}

// UpdateProduct - обновляет продукт
func (s *Storage) UpdateProduct(ctx context.Context, p models.Product) error {
	const fn = "storage.postgres.product.UpdateProduct"

	cmd, err := s.db.Exec(ctx,
		`UPDATE products SET name = $1, price = $2, stock = $3 WHERE id = $4`,
		p.Name, p.Price, p.Stock, p.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w: id=%d", fn, storage.ErrNotFound, p.ID)
	}

	return nil
}
