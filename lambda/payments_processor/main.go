package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Order represents an order from SNS
type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"`
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

// Item represents an item in an order
type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// handler processes SNS events containing orders
func handler(ctx context.Context, snsEvent events.SNSEvent) error {
	log.Printf("Received %d SNS messages\n", len(snsEvent.Records))

	for _, record := range snsEvent.Records {
		snsMessage := record.SNS
		log.Printf("Processing SNS message ID: %s\n", snsMessage.MessageID)

		// Parse order from SNS message
		var order Order
		if err := json.Unmarshal([]byte(snsMessage.Message), &order); err != nil {
			log.Printf("ERROR: Failed to unmarshal order: %v\n", err)
			// Return error to trigger retry
			return fmt.Errorf("failed to parse order: %w", err)
		}

		log.Printf("Processing order %s for customer %d with %d items\n",
			order.OrderID, order.CustomerID, len(order.Items))

		// Process the order (simulate 3-second payment processing)
		startTime := time.Now()
		if err := processPayment(order); err != nil {
			log.Printf("ERROR: Payment failed for order %s: %v\n", order.OrderID, err)
			return err // Trigger retry
		}

		processingTime := time.Since(startTime)
		log.Printf("Order %s completed successfully in %v\n", order.OrderID, processingTime)
	}

	return nil
}

// processPayment simulates payment processing with 3-second delay
func processPayment(order Order) error {
	log.Printf("Order %s: Processing payment...\n", order.OrderID)

	// Simulate 3-second payment processing (same as ECS version)
	time.Sleep(3 * time.Second)

	log.Printf("Order %s: Payment completed\n", order.OrderID)
	return nil
}

func main() {
	lambda.Start(handler)
}
