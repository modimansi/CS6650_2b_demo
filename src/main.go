package main

import (
	product "text/main/product"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	store := product.NewStore()
    store.SeedBulk(100000)
	handlers := product.NewHandlers(store)

	product.Register(router, handlers)

	router.Run(":8080")
}
