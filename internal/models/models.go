package models

import "time"

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Order struct {
	ID         int         `json:"id"`
	UserEmail  string      `json:"user_email"`
	TotalPrice float64     `json:"total_price"`
	CreatedAt  time.Time   `json:"created_at"`
	Items      []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID           int     `json:"id"`
	OrderID      int     `json:"order_id"`
	ProductID    int     `json:"product_id"`
	ProductName  string  `json:"product_name,omitempty"`
	ProductPrice float64 `json:"product_price,omitempty"`
	Quantity     int     `json:"quantity"`
}

type OrderDetail struct {
	OrderID           int         `json:"order_id"`
	UserEmail         string      `json:"user_email"`
	TotalPrice        float64     `json:"total_price"`
	CreatedAt         time.Time   `json:"created_at"`
	TransactionAmount float64     `json:"transaction_amount"`
	TransactionStatus string      `json:"transaction_status"`
	Items             []OrderItem `json:"items"`
}

type PopularProduct struct {
	ProductID     int     `json:"product_id"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Stock         int     `json:"stock"`
	TotalQuantity int64   `json:"total_quantity"`
}
