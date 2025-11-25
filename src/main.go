package main

import (
	"log"
	"os"
	"text/main/cart"
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

	// Initialize shopping cart handlers with database
	// Get cart store type from environment (postgres or dynamodb)
	storeType := os.Getenv("CART_STORE_TYPE")
	if storeType == "" {
		storeType = "postgres" // Default to PostgreSQL
	}

	var cartStore cart.CartStore
	var err error

	log.Printf("Initializing cart store with type: %s", storeType)

	if storeType == "dynamodb" {
		// Initialize DynamoDB store
		tableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if tableName == "" {
			tableName = "CS6650L2-shopping-carts" // Default table name
		}
		log.Printf("Using DynamoDB table: %s", tableName)
		cartStore, err = cart.NewDynamoDBStore(tableName)
	} else {
		// Initialize PostgreSQL store
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://postgres:postgres@localhost:5432/shopping?sslmode=disable"
			log.Println("DATABASE_URL not set, using default local connection")
		}
		cartStore, err = cart.NewStore(dbURL)
	}

	if err != nil {
		log.Printf("WARNING: Failed to initialize cart store: %v\n", err)
		log.Println("Cart endpoints will not be available")
	} else {
		// Initialize schema (optional, for PostgreSQL only)
		if initSchema := os.Getenv("INIT_DB_SCHEMA"); initSchema == "true" && storeType == "postgres" {
			if err := cartStore.InitSchema(); err != nil {
				log.Printf("WARNING: Failed to initialize database schema: %v\n", err)
			} else {
				log.Println("Database schema initialized successfully")
			}
		}

		cartHandlers := cart.NewHandlers(cartStore)
		cart.Register(router, cartHandlers)
		log.Printf("Shopping cart service initialized successfully with %s backend", storeType)
	}

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
