package main

import (
	"log"
	"text/main/orders"
	product "text/main/product"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Initialize product handlers
	store := product.NewStore()
	store.SeedBulk(100000)
	productHandlers := product.NewHandlers(store)
	product.Register(router, productHandlers)

	// Initialize order handlers
	orderHandlers := orders.NewHandlers()
	orders.Register(router, orderHandlers)

	// Start order processor (polls SQS and processes orders asynchronously)
	processor, err := orders.NewOrderProcessor()
	if err != nil {
		log.Printf("WARNING: Failed to initialize order processor: %v\n", err)
	} else if processor != nil {
		processor.Start()
		log.Println("Order processor started successfully")
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	log.Println("Starting server on :8080")
	router.Run(":8080")
}
