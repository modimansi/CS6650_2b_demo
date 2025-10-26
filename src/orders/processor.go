package orders

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// OrderProcessor continuously polls SQS and processes orders
type OrderProcessor struct {
	sqsClient *sqs.SQS
	queueURL  string
}

// NewOrderProcessor creates a new order processor
func NewOrderProcessor() (*OrderProcessor, error) {
	// Get SQS queue URL from environment
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Println("WARNING: SQS_QUEUE_URL not set, order processor will not start")
		return nil, nil
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return nil, err
	}

	// Create SQS client
	sqsClient := sqs.New(sess)

	return &OrderProcessor{
		sqsClient: sqsClient,
		queueURL:  queueURL,
	}, nil
}

// Start begins the order processing loop
func (p *OrderProcessor) Start() {
	if p == nil {
		log.Println("Order processor not initialized, skipping")
		return
	}

	log.Printf("Starting order processor, polling queue: %s\n", p.queueURL)

	// Run in a separate goroutine
	go p.pollLoop()
}

// pollLoop continuously polls SQS for messages
func (p *OrderProcessor) pollLoop() {
	for {
		// Receive messages from SQS (up to 10 messages, 20-second wait)
		result, err := p.sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(p.queueURL),
			MaxNumberOfMessages: aws.Int64(10), // Up to 10 messages
			WaitTimeSeconds:     aws.Int64(20), // Long polling (20 seconds)
			// Uses queue's default visibility timeout (30 seconds)
		})
		if err != nil {
			log.Printf("ERROR: Failed to receive messages from SQS: %v\n", err)
			time.Sleep(5 * time.Second) // Wait before retry
			continue
		}

		// Process each message in a separate goroutine
		for _, message := range result.Messages {
			go p.processMessage(message)
		}
	}
}

// processMessage processes a single order message
func (p *OrderProcessor) processMessage(message *sqs.Message) {
	log.Printf("Processing message: %s\n", *message.MessageId)

	// Extract SNS message body
	var snsMessage struct {
		Message string `json:"Message"`
	}
	if err := json.Unmarshal([]byte(*message.Body), &snsMessage); err != nil {
		log.Printf("ERROR: Failed to unmarshal SNS message: %v\n", err)
		// Still delete the message as it's malformed
		p.deleteMessage(message)
		return
	}

	// Parse order from SNS message
	var order Order
	if err := json.Unmarshal([]byte(snsMessage.Message), &order); err != nil {
		log.Printf("ERROR: Failed to unmarshal order: %v\n", err)
		// Delete malformed message
		p.deleteMessage(message)
		return
	}

	log.Printf("Processing order %s with %d items\n", order.OrderID, len(order.Items))

	// Process the order (includes 3-second payment delay)
	// This simulates payment processing with the same bottleneck as sync
	p.processOrder(order)

	// Delete message from queue after successful processing
	p.deleteMessage(message)

	log.Printf("Order %s completed and removed from queue\n", order.OrderID)
}

// processOrder simulates order processing with payment delay
func (p *OrderProcessor) processOrder(order Order) {
	// Acquire semaphore - blocks if another payment is processing
	// This maintains the same bottleneck as the sync endpoint
	paymentSemaphore <- struct{}{}
	defer func() { <-paymentSemaphore }()

	// Simulate 3-second payment processing
	log.Printf("Order %s: Processing payment...\n", order.OrderID)
	time.Sleep(3 * time.Second)
	log.Printf("Order %s: Payment completed\n", order.OrderID)
}

// deleteMessage removes a message from the SQS queue
func (p *OrderProcessor) deleteMessage(message *sqs.Message) {
	_, err := p.sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(p.queueURL),
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		log.Printf("ERROR: Failed to delete message %s: %v\n", *message.MessageId, err)
	}
}
