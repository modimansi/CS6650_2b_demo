package cart

import "time"

// ShoppingCart represents a shopping cart entity
// Supports both int (PostgreSQL) and string/UUID (DynamoDB) IDs
type ShoppingCart struct {
	ID         interface{} `json:"shopping_cart_id" db:"id"`
	CustomerID int         `json:"customer_id" db:"customer_id"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
}

// CartItem represents an item in a shopping cart
type CartItem struct {
	ID             int       `json:"id" db:"id"`
	ShoppingCartID int       `json:"shopping_cart_id" db:"shopping_cart_id"`
	ProductID      int       `json:"product_id" db:"product_id"`
	Quantity       int       `json:"quantity" db:"quantity"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// CartWithItems represents a shopping cart with all its items
type CartWithItems struct {
	ShoppingCart
	Items []CartItemDetail `json:"items"`
}

// CartItemDetail includes product information
// Supports both int (PostgreSQL) and string/UUID (DynamoDB) cart IDs
type CartItemDetail struct {
	ID             int         `json:"id"`
	ShoppingCartID interface{} `json:"shopping_cart_id"`
	ProductID      int         `json:"product_id"`
	ProductName    string      `json:"product_name,omitempty"`
	ProductPrice   float64     `json:"product_price,omitempty"`
	Quantity       int         `json:"quantity"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// CreateCartRequest represents the request to create a shopping cart
type CreateCartRequest struct {
	CustomerID int `json:"customer_id" binding:"required,min=1"`
}

// CreateCartResponse represents the response after creating a cart
// Supports both int (PostgreSQL) and string/UUID (DynamoDB) cart IDs
type CreateCartResponse struct {
	ShoppingCartID interface{} `json:"shopping_cart_id"`
}

// AddItemRequest represents the request to add items to a cart
type AddItemRequest struct {
	ProductID int `json:"product_id" binding:"required,min=1"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

// CheckoutResponse represents the response after checkout
// Supports both int (PostgreSQL) and string/UUID (DynamoDB) order IDs
type CheckoutResponse struct {
	OrderID interface{} `json:"order_id"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message"`
}
