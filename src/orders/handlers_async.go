package orders

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/gin-gonic/gin"
)

// POST /orders/async - Asynchronous order processing
func (h *Handlers) CreateOrderAsync(c *gin.Context) {
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

	// Get SNS topic ARN from environment
	snsTopicARN := os.Getenv("SNS_TOPIC_ARN")
	if snsTopicARN == "" {
		log.Println("ERROR: SNS_TOPIC_ARN environment variable not set")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "messaging service not configured"})
		return
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		log.Printf("ERROR: Failed to create AWS session: %v\n", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to initialize messaging service"})
		return
	}

	// Create SNS client
	snsClient := sns.New(sess)

	// Marshal order to JSON
	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Printf("ERROR: Failed to marshal order: %v\n", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to process order"})
		return
	}

	// Publish message to SNS
	_, err = snsClient.Publish(&sns.PublishInput{
		TopicArn: aws.String(snsTopicARN),
		Message:  aws.String(string(orderJSON)),
		Subject:  aws.String(fmt.Sprintf("Order %s", order.OrderID)),
	})
	if err != nil {
		log.Printf("ERROR: Failed to publish to SNS: %v\n", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to queue order for processing"})
		return
	}

	log.Printf("Order %s queued for async processing\n", order.OrderID)

	// Return 202 Accepted immediately
	c.JSON(http.StatusAccepted, gin.H{
		"order_id": order.OrderID,
		"status":   "queued",
		"message":  "Order queued for processing",
	})
}
