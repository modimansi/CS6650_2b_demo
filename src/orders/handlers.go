package orders

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Global semaphore - simulates payment processor capacity
// Size is configurable via WORKER_COUNT environment variable (default: 1)
// This creates the bottleneck needed to demonstrate async benefits
var paymentSemaphore chan struct{}

func init() {
	// Read worker count from environment (default to 1)
	workerCount := 1
	if envWorkers := os.Getenv("WORKER_COUNT"); envWorkers != "" {
		if count, err := strconv.Atoi(envWorkers); err == nil && count > 0 {
			workerCount = count
		}
	}

	// Initialize semaphore with worker count capacity
	paymentSemaphore = make(chan struct{}, workerCount)
	log.Printf("Payment processor initialized with %d concurrent workers\n", workerCount)
}

type Handlers struct{}

func NewHandlers() *Handlers {
	return &Handlers{}
}

// POST /orders/sync - Synchronous order processing
func (h *Handlers) CreateOrderSync(c *gin.Context) {
	var order Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid JSON body: " + err.Error()})
		return
	}

	// Validate order has items
	if len(order.Items) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "order must contain at least one item"})
		return
	}

	// Set created time if not provided
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now()
	}

	// Record start time for processing duration
	start := time.Now()

	// Simulate payment processing with 3-second delay using buffered channel
	// This approach uses a channel instead of just sleep
	paymentResult := h.processPaymentAsync(order)

	// Block waiting for payment processing to complete
	result := <-paymentResult

	processingTime := time.Since(start)

	// Check if payment was successful
	if !result.Success {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: result.Error})
		return
	}

	// Return success response
	response := OrderResponse{
		OrderID:        order.OrderID,
		Status:         "completed",
		ProcessingTime: processingTime.String(),
		Message:        "Order processed successfully",
	}

	c.JSON(http.StatusOK, response)
}

// PaymentResult represents the result of payment processing
type PaymentResult struct {
	Success bool
	Error   string
}

// processPaymentAsync simulates payment processing using a semaphore to create a real bottleneck
// The semaphore (buffered channel with size 1) ensures only 1 payment can process at a time
// This creates the bottleneck that causes failures during flash sale scenarios
func (h *Handlers) processPaymentAsync(order Order) <-chan PaymentResult {
	// Create a buffered channel with capacity of 1 for the result
	resultChan := make(chan PaymentResult, 1)

	// Spawn a goroutine to simulate payment processing
	go func() {
		// CRITICAL: Acquire semaphore - blocks if another payment is processing
		// This simulates a single-threaded payment processor bottleneck
		paymentSemaphore <- struct{}{}

		// Ensure we release the semaphore when done
		defer func() {
			<-paymentSemaphore // Release the semaphore
		}()

		// Now do the 3-second payment processing
		// Only ONE request can be here at a time due to the semaphore
		timer := time.NewTimer(3 * time.Second)
		<-timer.C

		// Simulate payment processing logic
		// In a real system, this would call a payment gateway
		result := PaymentResult{
			Success: true,
			Error:   "",
		}

		// Send result through the buffered channel
		resultChan <- result
	}()

	return resultChan
}

// Alternative implementation using time.After if preferred:
// func (h *Handlers) processPaymentWithAfter(order Order) <-chan PaymentResult {
//     resultChan := make(chan PaymentResult, 1)
//
//     go func() {
//         // Wait for 3 seconds using time.After channel
//         <-time.After(3 * time.Second)
//
//         resultChan <- PaymentResult{Success: true}
//     }()
//
//     return resultChan
// }

// Health check endpoint
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "orders",
	})
}
