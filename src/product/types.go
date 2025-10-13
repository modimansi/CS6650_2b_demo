package product

// Product represents a product entity.
type Product struct {
	ID          int32   `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category,omitempty"`
	Description string  `json:"description,omitempty"`
	Brand       string  `json:"brand,omitempty"`
	Price       float64 `json:"price,omitempty"`
}

// ErrorResponse is a basic error payload.
type ErrorResponse struct {
	Message string `json:"message"`
}

// SearchResponse is the response envelope for limited searches.
type SearchResponse struct {
	Products   []Product `json:"products"`
	TotalFound int       `json:"total_found"`
	SearchTime string    `json:"search_time,omitempty"`
}
