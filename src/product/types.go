package product

// Product represents a product entity.
type Product struct {
	ID          int32   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price,omitempty"`
}

// ErrorResponse is a basic error payload.
type ErrorResponse struct {
	Message string `json:"message"`
}
