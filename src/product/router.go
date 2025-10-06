package product

import "github.com/gin-gonic/gin"

// Register mounts product routes on the provided router group or engine.
func Register(r gin.IRoutes, h *Handlers) {
	r.POST("/products", h.CreateProduct)
	r.GET("/products/:productId", h.GetProduct)
	r.POST("/products/:productId/details", h.AddProductDetails)
}
