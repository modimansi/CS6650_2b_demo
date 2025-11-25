package cart

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// DynamoDBStore handles shopping cart operations with DynamoDB
type DynamoDBStore struct {
	client    *dynamodb.Client
	tableName string
}

// DynamoDBCart represents a cart item in DynamoDB
type DynamoDBCart struct {
	CartID         string             `dynamodbav:"cart_id"`
	CustomerID     int                `dynamodbav:"customer_id"`
	Items          []DynamoDBCartItem `dynamodbav:"items"`
	ItemCount      int                `dynamodbav:"item_count"`
	TotalAmount    float64            `dynamodbav:"total_amount"`
	CreatedAt      string             `dynamodbav:"created_at"`
	UpdatedAt      string             `dynamodbav:"updated_at"`
	ExpirationTime int64              `dynamodbav:"expiration_time"`
}

// DynamoDBCartItem represents an item within a cart
type DynamoDBCartItem struct {
	ProductID    int     `dynamodbav:"product_id"`
	Quantity     int     `dynamodbav:"quantity"`
	ProductName  string  `dynamodbav:"product_name"`
	ProductPrice float64 `dynamodbav:"product_price"`
}

// NewDynamoDBStore creates a new DynamoDB store
func NewDynamoDBStore(tableName string) (*DynamoDBStore, error) {
	ctx := context.TODO()

	// Load AWS SDK config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

	return &DynamoDBStore{
		client:    client,
		tableName: tableName,
	}, nil
}

// CreateCart creates a new shopping cart in DynamoDB
func (s *DynamoDBStore) CreateCart(customerID int) (*ShoppingCart, error) {
	ctx := context.TODO()

	// Generate UUID for cart_id (ensures even distribution)
	cartID := uuid.New().String()
	now := time.Now()

	// Create DynamoDB item
	cart := DynamoDBCart{
		CartID:         cartID,
		CustomerID:     customerID,
		Items:          []DynamoDBCartItem{},
		ItemCount:      0,
		TotalAmount:    0.0,
		CreatedAt:      now.UTC().Format(time.RFC3339),
		UpdatedAt:      now.UTC().Format(time.RFC3339),
		ExpirationTime: now.Add(7 * 24 * time.Hour).Unix(), // TTL: 7 days
	}

	// Marshal to DynamoDB attribute values
	item, err := attributevalue.MarshalMap(cart)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cart: %w", err)
	}

	// Put item in DynamoDB
	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create cart in DynamoDB: %w", err)
	}

	// Return cart in format expected by API
	return &ShoppingCart{
		ID:         cartID,
		CustomerID: customerID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// GetCart retrieves a shopping cart by ID (no items)
func (s *DynamoDBStore) GetCart(cartIDInterface interface{}) (*ShoppingCart, error) {
	// Convert interface{} to string (UUID format)
	cartID, ok := cartIDInterface.(string)
	if !ok {
		return nil, errors.New("invalid cart ID type for DynamoDB (expected string UUID)")
	}

	ctx := context.TODO()

	// Get item from DynamoDB
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"cart_id": &types.AttributeValueMemberS{Value: cartID},
		},
		ConsistentRead: aws.Bool(false), // Eventual consistency for performance
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get cart from DynamoDB: %w", err)
	}

	// Check if item exists
	if result.Item == nil {
		return nil, ErrCartNotFound
	}

	// Unmarshal DynamoDB item
	var cart DynamoDBCart
	err = attributevalue.UnmarshalMap(result.Item, &cart)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart: %w", err)
	}

	// Parse timestamps
	createdAt, _ := time.Parse(time.RFC3339, cart.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, cart.UpdatedAt)

	return &ShoppingCart{
		ID:         cart.CartID,
		CustomerID: cart.CustomerID,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

// GetCartWithItems retrieves a cart with all its items
func (s *DynamoDBStore) GetCartWithItems(cartIDInterface interface{}) (*CartWithItems, error) {
	// Convert interface{} to string (UUID format)
	cartID, ok := cartIDInterface.(string)
	if !ok {
		return nil, errors.New("invalid cart ID type for DynamoDB (expected string UUID)")
	}

	ctx := context.TODO()

	// Get item from DynamoDB with eventual consistency
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"cart_id": &types.AttributeValueMemberS{Value: cartID},
		},
		ConsistentRead: aws.Bool(false), // Eventual consistency
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get cart from DynamoDB: %w", err)
	}

	// Check if item exists
	if result.Item == nil {
		return nil, ErrCartNotFound
	}

	// Unmarshal DynamoDB item
	var cart DynamoDBCart
	err = attributevalue.UnmarshalMap(result.Item, &cart)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart: %w", err)
	}

	// Parse timestamps
	createdAt, _ := time.Parse(time.RFC3339, cart.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, cart.UpdatedAt)

	// Convert DynamoDB items to API format
	items := make([]CartItemDetail, len(cart.Items))
	for i, item := range cart.Items {
		items[i] = CartItemDetail{
			ID:             i + 1, // Generate sequential ID for consistency with PostgreSQL API
			ShoppingCartID: cartID,
			ProductID:      item.ProductID,
			ProductName:    item.ProductName,
			ProductPrice:   item.ProductPrice,
			Quantity:       item.Quantity,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		}
	}

	return &CartWithItems{
		ShoppingCart: ShoppingCart{
			ID:         cart.CartID,
			CustomerID: cart.CustomerID,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		},
		Items: items,
	}, nil
}

// AddOrUpdateItem adds or updates an item in the cart
func (s *DynamoDBStore) AddOrUpdateItem(cartIDInterface interface{}, productID int, quantity int) error {
	// Convert interface{} to string (UUID format)
	cartID, ok := cartIDInterface.(string)
	if !ok {
		return errors.New("invalid cart ID type for DynamoDB (expected string UUID)")
	}

	ctx := context.TODO()

	// First, verify cart exists and get current items
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"cart_id": &types.AttributeValueMemberS{Value: cartID},
		},
		ConsistentRead: aws.Bool(true), // Strong consistency for updates
	})

	if err != nil {
		return fmt.Errorf("failed to get cart: %w", err)
	}

	if result.Item == nil {
		return ErrCartNotFound
	}

	// Unmarshal current cart
	var cart DynamoDBCart
	err = attributevalue.UnmarshalMap(result.Item, &cart)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cart: %w", err)
	}

	// Mock product data (in real app, would fetch from products table)
	// For testing purposes, we use the same product seeding logic
	productName := fmt.Sprintf("Product %d", productID)
	productPrice := float64((productID%100)+1) + float64(productID%100)/100.0

	// Check if product already exists in cart
	found := false
	for i, item := range cart.Items {
		if item.ProductID == productID {
			// Update quantity
			cart.Items[i].Quantity += quantity
			found = true
			break
		}
	}

	// If not found, add new item
	if !found {
		cart.Items = append(cart.Items, DynamoDBCartItem{
			ProductID:    productID,
			Quantity:     quantity,
			ProductName:  productName,
			ProductPrice: productPrice,
		})
	}

	// Recalculate totals
	cart.ItemCount = 0
	cart.TotalAmount = 0.0
	for _, item := range cart.Items {
		cart.ItemCount += item.Quantity
		cart.TotalAmount += float64(item.Quantity) * item.ProductPrice
	}

	// Update timestamp
	cart.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	// Marshal updated cart
	item, err := attributevalue.MarshalMap(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	// Update item in DynamoDB
	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to update cart in DynamoDB: %w", err)
	}

	return nil
}

// CheckoutCart processes checkout (placeholder for DynamoDB)
func (s *DynamoDBStore) CheckoutCart(cartIDInterface interface{}) (interface{}, error) {
	// Convert interface{} to string (UUID format)
	cartID, ok := cartIDInterface.(string)
	if !ok {
		return "", errors.New("invalid cart ID type for DynamoDB (expected string UUID)")
	}

	ctx := context.TODO()

	// Get cart
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"cart_id": &types.AttributeValueMemberS{Value: cartID},
		},
		ConsistentRead: aws.Bool(true), // Strong consistency for checkout
	})

	if err != nil {
		return "", fmt.Errorf("failed to get cart: %w", err)
	}

	if result.Item == nil {
		return "", ErrCartNotFound
	}

	// Unmarshal cart
	var cart DynamoDBCart
	err = attributevalue.UnmarshalMap(result.Item, &cart)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal cart: %w", err)
	}

	// Validate cart has items
	if len(cart.Items) == 0 {
		return "", ErrEmptyCart
	}

	// Generate order ID (in real app, would create order in database)
	orderID := uuid.New().String()

	// Delete cart after successful checkout
	_, err = s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"cart_id": &types.AttributeValueMemberS{Value: cartID},
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to delete cart after checkout: %w", err)
	}

	return orderID, nil
}

// Close is a no-op for DynamoDB (no connection to close)
func (s *DynamoDBStore) Close() error {
	return nil
}

// InitSchema is not needed for DynamoDB (table created by Terraform)
func (s *DynamoDBStore) InitSchema() error {
	return nil
}
