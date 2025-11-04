package cart

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handlers contains HTTP handlers for shopping cart operations
type Handlers struct {
	store *Store
}

// NewHandlers creates a new Handlers instance with the given store
func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

// NewHandlersWithDB creates a new Handlers instance with a database connection
func NewHandlersWithDB(db *sql.DB) *Handlers {
	store := NewStoreWithDB(db)
	return &Handlers{store: store}
}

// CreateCart handles POST /shopping-carts
// Creates a new shopping cart for a customer
func (h *Handlers) CreateCart(c *gin.Context) {
	var req CreateCartRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Additional validation
	if req.CustomerID < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "customer_id must be greater than 0",
		})
		return
	}

	// Create shopping cart
	cart, err := h.store.CreateCart(req.CustomerID)
	if err != nil {
		log.Printf("Error creating cart: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Failed to create shopping cart",
		})
		return
	}

	// Return success response with 201 Created
	c.JSON(http.StatusCreated, CreateCartResponse{
		ShoppingCartID: cart.ID,
	})
}

// GetCart handles GET /shopping-carts/:shoppingCartId
// Retrieves a cart with all its items using efficient JOINs
func (h *Handlers) GetCart(c *gin.Context) {
	// Parse cart ID from path parameter
	cartIDStr := c.Param("shoppingCartId")
	cartID, err := strconv.Atoi(cartIDStr)
	if err != nil || cartID < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid shopping cart ID",
		})
		return
	}

	// Retrieve cart with items
	cartWithItems, err := h.store.GetCartWithItems(cartID)
	if err != nil {
		// Distinguish between not found and server errors
		if errors.Is(err, ErrCartNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Message: "Shopping cart not found",
			})
			return
		}

		log.Printf("Error retrieving cart %d: %v", cartID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Failed to retrieve shopping cart",
		})
		return
	}

	// Return cart with items
	c.JSON(http.StatusOK, cartWithItems)
}

// AddItemToCart handles POST /shopping-carts/:shoppingCartId/items
// Adds or updates items in an existing cart
func (h *Handlers) AddItemToCart(c *gin.Context) {
	// Parse cart ID from path parameter
	cartIDStr := c.Param("shoppingCartId")
	cartID, err := strconv.Atoi(cartIDStr)
	if err != nil || cartID < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid shopping cart ID",
		})
		return
	}

	// Parse and validate request body
	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Additional validation
	if req.ProductID < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "product_id must be greater than 0",
		})
		return
	}
	if req.Quantity < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "quantity must be greater than 0",
		})
		return
	}

	// Add or update item in cart
	err = h.store.AddOrUpdateItem(cartID, req.ProductID, req.Quantity)
	if err != nil {
		// Handle specific errors appropriately
		if errors.Is(err, ErrCartNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Message: "Shopping cart not found",
			})
			return
		}
		if errors.Is(err, ErrProductNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Message: "Product not found",
			})
			return
		}

		log.Printf("Error adding item to cart %d: %v", cartID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Failed to add item to cart",
		})
		return
	}

	// Return 204 No Content on success
	c.Status(http.StatusNoContent)
}

// CheckoutCart handles POST /shopping-carts/:shoppingCartId/checkout
// Processes checkout for a shopping cart
func (h *Handlers) CheckoutCart(c *gin.Context) {
	// Parse cart ID from path parameter
	cartIDStr := c.Param("shoppingCartId")
	cartID, err := strconv.Atoi(cartIDStr)
	if err != nil || cartID < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid shopping cart ID",
		})
		return
	}

	// Process checkout
	orderID, err := h.store.CheckoutCart(cartID)
	if err != nil {
		// Handle specific errors appropriately
		if errors.Is(err, ErrCartNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Message: "Shopping cart not found",
			})
			return
		}
		if errors.Is(err, ErrEmptyCart) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "Cannot checkout an empty cart",
			})
			return
		}

		log.Printf("Error checking out cart %d: %v", cartID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Failed to process checkout",
		})
		return
	}

	// Return order ID with 200 OK
	c.JSON(http.StatusOK, CheckoutResponse{
		OrderID: orderID,
	})
}

// HealthCheck provides a health check endpoint for the cart service
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "shopping-cart",
	})
}
