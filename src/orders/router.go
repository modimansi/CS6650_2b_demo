package orders

import "github.com/gin-gonic/gin"

// Register mounts order routes on the provided router group or engine.
func Register(r gin.IRoutes, h *Handlers) {
	r.POST("/orders/sync", h.CreateOrderSync)
	r.POST("/orders/async", h.CreateOrderAsync)
}
