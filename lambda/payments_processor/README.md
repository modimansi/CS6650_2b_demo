# Lambda Payment Processor

This Lambda function processes orders from SNS with the same 3-second payment delay as the ECS processor.

## Architecture

```
POST /orders/async (ECS)
        ↓
    Publish to SNS
        ↓
    SNS Topic
        ↓
    Lambda Function (triggered directly)
        ↓
    Process payment (3 seconds)
```

**No SQS queue needed!** Lambda is invoked directly by SNS.

## Build

### Linux/Mac
```bash
chmod +x build.sh
./build.sh
```

### Windows
```powershell
.\build.ps1
```

This creates `function.zip` ready for deployment.

## Deploy

```bash
cd ../../terraform
terraform apply
```

## Test

### Test Single Order
```bash
# Get SNS topic ARN
SNS_TOPIC=$(cd terraform && terraform output -raw sns_topic_arn)

# Publish test order
aws sns publish \
    --topic-arn $SNS_TOPIC \
    --message '{
        "order_id": "TEST-001",
        "customer_id": 12345,
        "status": "pending",
        "items": [
            {
                "product_id": "PROD-001",
                "quantity": 1,
                "price": 29.99
            }
        ],
        "created_at": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
    }'
```

### View Logs
```bash
aws logs tail /aws/lambda/CS6650L2-payment-processor --follow
```

### Run Load Test
```bash
cd testing
locust -f orders_async_locustfile.py \
    --host=http://YOUR-IP:8080 \
    --users 100 --spawn-rate 20 --run-time 60s --headless
```

## Key Features

- **Automatic Scaling**: AWS scales Lambda concurrency automatically (up to 1000)
- **Built-in Retries**: 3 automatic retries on failure
- **No Queue Management**: SNS → Lambda directly
- **Pay Per Use**: Only pay for actual invocations
- **Same Processing Logic**: 3-second payment delay

## Comparison vs ECS

| Feature | ECS | Lambda |
|---------|-----|--------|
| Scaling | Manual (worker_count) | Automatic |
| Queue | Need SQS | Not needed |
| Polling | You implement | AWS handles |
| Retries | You implement | Built-in (3x) |
| Cold Start | None | ~100-200ms |
| Cost | $12/month + usage | Pay per invocation |
| Ops Burden | High | Zero |

## Configuration

- **Memory**: 512 MB (configurable in terraform/variables.tf)
- **Timeout**: 10 seconds (enough for 3s processing + overhead)
- **Runtime**: Go (provided.al2)
- **Trigger**: SNS topic subscription

## Monitoring

### CloudWatch Metrics
```bash
# Invocations
aws cloudwatch get-metric-statistics \
    --namespace AWS/Lambda \
    --metric-name Invocations \
    --dimensions Name=FunctionName,Value=CS6650L2-payment-processor \
    --start-time $(date -u -d '10 minutes ago' +%Y-%m-%dT%H:%M:%S) \
    --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
    --period 60 \
    --statistics Sum

# Errors
aws cloudwatch get-metric-statistics \
    --namespace AWS/Lambda \
    --metric-name Errors \
    --dimensions Name=FunctionName,Value=CS6650L2-payment-processor \
    --start-time $(date -u -d '10 minutes ago' +%Y-%m-%dT%H:%M:%S) \
    --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
    --period 60 \
    --statistics Sum

# Duration
aws cloudwatch get-metric-statistics \
    --namespace AWS/Lambda \
    --metric-name Duration \
    --dimensions Name=FunctionName,Value=CS6650L2-payment-processor \
    --start-time $(date -u -d '10 minutes ago' +%Y-%m-%dT%H:%M:%S) \
    --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
    --period 60 \
    --statistics Average
```

## Expected Behavior

During load test (100 users, 60s):
- **Orders accepted**: ~6,000 (same as before)
- **Lambda invocations**: ~6,000 (one per order)
- **Concurrent executions**: Auto-scales to match load
- **Processing time**: 3 seconds per order
- **No queue buildup**: Lambda scales automatically

## Files

```
lambda/payments_processor/
├── main.go          # Lambda handler
├── go.mod           # Go dependencies
├── go.sum           # Dependency checksums
├── build.sh         # Build script (Linux/Mac)
├── build.ps1        # Build script (Windows)
├── README.md        # This file
└── function.zip     # Deployment package (generated)
```

