package cart

import "github.com/gin-gonic/gin"

// Register mounts shopping cart routes on the provided router
func Register(r gin.IRoutes, h *Handlers) {
	// Create a new shopping cart
	r.POST("/shopping-carts", h.CreateCart)

	// Get shopping cart with items
	r.GET("/shopping-carts/:shoppingCartId", h.GetCart)

	// Add items to shopping cart
	r.POST("/shopping-carts/:shoppingCartId/items", h.AddItemToCart)

	// Checkout shopping cart
	r.POST("/shopping-carts/:shoppingCartId/checkout", h.CheckoutCart)
}
