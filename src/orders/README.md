# Orders API - Synchronous Processing

## Overview

This module implements synchronous order processing with simulated payment gateway integration.

## Endpoints

### POST /orders/sync

Synchronously processes an order with payment processing simulation.

**Key Features:**
- Accepts order data via JSON POST request
- Simulates 3-second payment processing using buffered channels (not sleep)
- Returns response only after payment processing completes
- Blocks the HTTP connection until processing is done

**Request Body:**
```json
{
  "order_id": "ORD-12345",
  "customer_id": 67890,
  "status": "pending",
  "items": [
    {
      "product_id": "PROD-001",
      "quantity": 2,
      "price": 29.99
    },
    {
      "product_id": "PROD-002",
      "quantity": 1,
      "price": 49.99
    }
  ],
  "created_at": "2025-10-25T10:30:00Z"
}
```

**Response (200 OK):**
```json
{
  "order_id": "ORD-12345",
  "status": "completed",
  "processing_time": "3.001234567s",
  "message": "Order processed successfully"
}
```

**Error Response (400 Bad Request):**
```json
{
  "message": "invalid JSON body: ..."
}
```

## Implementation Details

### Synchronous Processing with Buffered Channels

The payment processing uses a **buffered channel** approach instead of simple `time.Sleep()`:

```go
func (h *Handlers) processPaymentAsync(order Order) <-chan PaymentResult {
    // Create a buffered channel with capacity of 1
    resultChan := make(chan PaymentResult, 1)

    // Spawn a goroutine to simulate payment processing
    go func() {
        // Create a timer channel that will send after 3 seconds
        timer := time.NewTimer(3 * time.Second)
        
        // Wait for the timer
        <-timer.C

        // Process payment and send result
        resultChan <- PaymentResult{Success: true}
    }()

    return resultChan
}
```

**Why buffered channel?**
- Demonstrates proper Go concurrency patterns
- Prevents goroutine blocking if receiver isn't ready
- Allows async work while maintaining sync API behavior
- Shows channel-based communication vs simple blocking

### Processing Flow

```
1. HTTP Request arrives
   ↓
2. Validate JSON payload
   ↓
3. Start payment processing (goroutine)
   ↓
4. Block on channel receive (<-paymentResult)
   ↓
5. Wait 3 seconds (timer channel)
   ↓
6. Receive payment result
   ↓
7. Return HTTP response
```

## Testing

### Using curl

```bash
# Test successful order
curl -X POST http://localhost:8080/orders/sync \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "ORD-12345",
    "customer_id": 67890,
    "status": "pending",
    "items": [
      {
        "product_id": "PROD-001",
        "quantity": 2,
        "price": 29.99
      }
    ]
  }'
```

### Using httpie

```bash
# Install httpie: pip install httpie

http POST http://localhost:8080/orders/sync \
  order_id="ORD-12345" \
  customer_id:=67890 \
  status="pending" \
  items:='[{"product_id":"PROD-001","quantity":2,"price":29.99}]'
```

### Testing with Load (Locust)

```python
from locust import HttpUser, task, between

class OrderUser(HttpUser):
    wait_time = between(1, 3)
    
    @task
    def create_order_sync(self):
        self.client.post("/orders/sync", json={
            "order_id": f"ORD-{self.environment.runner.user_count}",
            "customer_id": 12345,
            "status": "pending",
            "items": [
                {
                    "product_id": "PROD-001",
                    "quantity": 1,
                    "price": 29.99
                }
            ]
        })
```

## Performance Characteristics

### Synchronous Processing

- **Latency**: Minimum 3 seconds per request (payment processing delay)
- **Throughput**: Limited by connection pool and processing time
- **Connection**: Blocks HTTP connection for entire duration
- **Scalability**: Limited - each request holds a connection for 3+ seconds

### Expected Behavior

With synchronous processing:
- 1 request takes ~3 seconds
- 10 concurrent requests take ~3 seconds (if sufficient connections)
- 100 concurrent requests may exhaust connection pool
- Client waits for full processing before receiving response

## Health Check

The `/health` endpoint is available at the application root level for load balancer health checks:

```bash
curl http://localhost:8080/health
# Response: ok
```

## Comparison: Sync vs Async (Future)

| Aspect | Synchronous (/orders/sync) | Asynchronous (Future) |
|--------|---------------------------|----------------------|
| Response Time | 3+ seconds | Immediate (< 100ms) |
| Client Blocking | Yes | No |
| Scalability | Limited | High |
| Implementation | Simple | Complex (queues, workers) |
| Use Case | Real-time confirmation | High throughput |

## Code Structure

```
src/orders/
├── types.go       # Order, Item, Response structs
├── handlers.go    # HTTP handlers and payment simulation
├── router.go      # Route registration
└── README.md      # This file
```

## Next Steps (Phase 2)

The asynchronous implementation will:
1. Accept order immediately
2. Return 202 Accepted with tracking ID
3. Process payment asynchronously via SQS queue
4. Use Lambda function for payment processing
5. Update order status via callback/webhook

