package orders

import "time"

// Order represents an order entity.
type Order struct {
	OrderID    string    `json:"order_id" binding:"required"`
	CustomerID int       `json:"customer_id" binding:"required"`
	Status     string    `json:"status" binding:"required"`
	Items      []Item    `json:"items" binding:"required"`
	CreatedAt  time.Time `json:"created_at"`
}

// Item represents an item within an order.
type Item struct {
	ProductID string  `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	Price     float64 `json:"price" binding:"required,min=0"`
}

// OrderResponse is returned after successful order processing.
type OrderResponse struct {
	OrderID        string `json:"order_id"`
	Status         string `json:"status"`
	ProcessingTime string `json:"processing_time"`
	Message        string `json:"message"`
}

// ErrorResponse is a basic error payload.
type ErrorResponse struct {
	Message string `json:"message"`
}
