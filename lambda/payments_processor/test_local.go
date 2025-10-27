//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

// Copy the handler and processing functions from main.go
type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"`
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

func processPayment(order Order) error {
	log.Printf("Order %s: Processing payment...\n", order.OrderID)
	time.Sleep(3 * time.Second)
	log.Printf("Order %s: Payment completed\n", order.OrderID)
	return nil
}

func handler(ctx context.Context, snsEvent events.SNSEvent) error {
	log.Printf("Received %d SNS messages\n", len(snsEvent.Records))

	for _, record := range snsEvent.Records {
		snsMessage := record.SNS
		log.Printf("Processing SNS message ID: %s\n", snsMessage.MessageID)

		var order Order
		if err := json.Unmarshal([]byte(snsMessage.Message), &order); err != nil {
			log.Printf("ERROR: Failed to unmarshal order: %v\n", err)
			return fmt.Errorf("failed to parse order: %w", err)
		}

		log.Printf("Processing order %s for customer %d with %d items\n",
			order.OrderID, order.CustomerID, len(order.Items))

		startTime := time.Now()
		if err := processPayment(order); err != nil {
			log.Printf("ERROR: Payment failed for order %s: %v\n", order.OrderID, err)
			return err
		}

		processingTime := time.Since(startTime)
		log.Printf("Order %s completed successfully in %v\n", order.OrderID, processingTime)
	}

	return nil
}

func main() {
	// Create test SNS event
	testOrder := Order{
		OrderID:    "LOCAL-TEST-001",
		CustomerID: 12345,
		Status:     "pending",
		Items: []Item{
			{
				ProductID: "PROD-001",
				Quantity:  2,
				Price:     29.99,
			},
		},
		CreatedAt: time.Now(),
	}

	orderJSON, _ := json.Marshal(testOrder)

	snsEvent := events.SNSEvent{
		Records: []events.SNSEventRecord{
			{
				SNS: events.SNSEntity{
					MessageID: "local-test-message",
					Message:   string(orderJSON),
				},
			},
		},
	}

	// Run handler
	log.Println("🧪 Testing Lambda Handler Locally")
	log.Println("===================================")

	ctx := context.Background()
	if err := handler(ctx, snsEvent); err != nil {
		log.Fatalf("Handler failed: %v", err)
	}

	log.Println("===================================")
	log.Println("✅ Test completed successfully!")
}
