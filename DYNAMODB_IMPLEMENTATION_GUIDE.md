# DynamoDB Shopping Cart Implementation Guide

## 🎯 Overview

This guide documents the implementation of DynamoDB as a NoSQL alternative to PostgreSQL for shopping cart storage, enabling performance comparison between relational and NoSQL databases.

---

## 📋 What Was Implemented

### 1. **Terraform Infrastructure** ✅

**Location:** `terraform/modules/dynamodb/`

**Created Files:**
- `main.tf` - DynamoDB table configuration
- `variables.tf` - Module variables
- `outputs.tf` - Module outputs
- `DESIGN_DECISIONS.md` - Complete design rationale (15+ pages)

**Key Features:**
- DynamoDB table with PAY_PER_REQUEST billing
- Global Secondary Index for customer lookups
- TTL for automatic cart expiration (7 days)
- Point-in-time recovery (optional)
- Server-side encryption enabled

### 2. **Integration with Existing Infrastructure** ✅

**Updated Files:**
- `terraform/main.tf` - Added DynamoDB module
- `terraform/variables.tf` - Added DynamoDB variables
- `terraform/outputs.tf` - Added DynamoDB outputs
- `terraform/modules/ecs/main.tf` - Added environment variables
- `terraform/modules/ecs/variables.tf` - Added DynamoDB table name variable

**Environment Variables Added to ECS:**
```bash
DYNAMODB_TABLE_NAME=CS6650L2-shopping-carts
CART_STORE_TYPE=postgres  # Can switch to "dynamodb" later
```

---

## 🏗️ DynamoDB Schema Design

### Table Structure

```
Table Name: CS6650L2-shopping-carts
Partition Key: cart_id (String - UUID)
Sort Key: None
```

### Item Structure

```json
{
  "cart_id": "550e8400-e29b-41d4-a716-446655440000",
  "customer_id": 12345,
  "items": [
    {
      "product_id": 1,
      "quantity": 2,
      "product_name": "Widget Alpha",
      "product_price": 29.99
    },
    {
      "product_id": 50,
      "quantity": 1,
      "product_name": "Gadget Beta",
      "product_price": 49.99
    }
  ],
  "item_count": 3,
  "total_amount": 109.97,
  "created_at": "2025-01-19T22:00:00Z",
  "updated_at": "2025-01-19T22:15:00Z",
  "expiration_time": 1738195200
}
```

### Global Secondary Index

```
Index Name: CustomerIndex
Partition Key: customer_id (Number)
Projection: ALL
```

---

## 🎨 Design Decisions Summary

### Why cart_id as Partition Key?

✅ **Even distribution** - UUIDs spread evenly across partitions  
✅ **No hot partitions** - Each cart independent  
✅ **Natural access pattern** - Primary operation is "get cart by ID"  
✅ **Scalable** - Works with millions of carts

### Why Embed Items vs Separate Table?

✅ **Single read operation** - Get entire cart in one query  
✅ **Cost efficient** - 1 RCU vs N RCUs  
✅ **Atomic updates** - Cart and items updated together  
✅ **<50ms requirement** - Easily met with single item read

### Why PAY_PER_REQUEST?

✅ **Variable load** - Testing creates burst traffic  
✅ **No capacity planning** - Auto-scales instantly  
✅ **Cost efficient** - Only pay for actual operations  
✅ **Assignment friendly** - No need to predict traffic

### Key Trade-offs

| Decision | Pro | Con |
|----------|-----|-----|
| Embed items | Fast reads, low cost | 400KB item limit |
| cart_id PK | Even distribution | Need GSI for customer queries |
| Eventual consistency | Lower cost, faster | Slight delay possible |
| 7-day TTL | Auto-cleanup, cost savings | Can't recover abandoned carts |

---

## 🚀 Deployment Instructions

### Step 1: Deploy DynamoDB Table

```powershell
cd terraform

# Initialize new module
terraform init

# Review changes (should see DynamoDB table creation)
terraform plan

# Deploy
terraform apply -auto-approve
```

**Expected Output:**
```
module.dynamodb.aws_dynamodb_table.shopping_carts: Creating...
module.dynamodb.aws_dynamodb_table.shopping_carts: Creation complete [1s]

Outputs:
dynamodb_table_name = "CS6650L2-shopping-carts"
dynamodb_table_arn = "arn:aws:dynamodb:us-west-2:...:table/CS6650L2-shopping-carts"
```

### Step 2: Verify DynamoDB Table

```powershell
# Check table exists
terraform output dynamodb_table_name

# View in AWS Console
# https://us-west-2.console.aws.amazon.com/dynamodbv2/home?region=us-west-2#tables
```

### Step 3: Implementation Status

Currently deployed with **PostgreSQL active** (`CART_STORE_TYPE=postgres`)

**Next Steps Required:**
1. ⏳ Implement Go DynamoDB store (`src/cart/store_dynamodb.go`)
2. ⏳ Update router to support both databases
3. ⏳ Add environment variable to switch: `CART_STORE_TYPE=dynamodb`
4. ⏳ Create comparison test script
5. ⏳ Run performance tests

---

## 📊 Expected Performance Comparison

### Predictions

| Operation | PostgreSQL (Actual) | DynamoDB (Predicted) | Expected Winner |
|-----------|--------------------|--------------------|-----------------|
| Create cart | 94.47ms avg | 15-25ms | ⚡ DynamoDB (4x faster) |
| Get cart | 42.14ms avg | 5-15ms | ⚡ DynamoDB (3x faster) |
| Add items | N/A (failed) | 15-25ms | ⚡ DynamoDB |
| Customer history | ~20ms | 20-50ms (GSI) | 🤝 Similar |

### Why DynamoDB Should Be Faster

1. **No JOINs** - Single item read vs 3-table JOIN
2. **No connection pool** - HTTP API vs TCP connections
3. **Optimized for key-value** - Primary design goal
4. **Managed service** - AWS handles all infrastructure
5. **SSD-backed** - Fast I/O by default

### Potential PostgreSQL Advantages

1. **Complex queries** - Full SQL support
2. **ACID transactions** - Strong consistency guarantees
3. **Mature ecosystem** - More tools and libraries
4. **Cost at scale** - Fixed instance cost vs per-request pricing

---

## 🧪 Testing Strategy

### Phase 1: Infrastructure Testing (Current Status)

```powershell
# Verify table exists
aws dynamodb describe-table --table-name CS6650L2-shopping-carts

# Check IAM permissions
aws dynamodb scan --table-name CS6650L2-shopping-carts --limit 1
```

### Phase 2: Implementation Testing (Pending)

After Go implementation:

```powershell
# Test DynamoDB endpoints
export CART_STORE_TYPE=dynamodb

# Create cart
curl -X POST http://$PUBLIC_IP:8080/shopping-carts \
  -H "Content-Type: application/json" \
  -d '{"customer_id": 1}'

# Add items
curl -X POST http://$PUBLIC_IP:8080/shopping-carts/$CART_ID/items \
  -H "Content-Type: application/json" \
  -d '{"product_id": 1, "quantity": 2}'

# Get cart
curl http://$PUBLIC_IP:8080/shopping-carts/$CART_ID
```

### Phase 3: Performance Comparison (Pending)

```powershell
cd testing

# Test PostgreSQL
python cart_performance_test.py --host http://$PUBLIC_IP:8080 --output postgres_results.json

# Switch to DynamoDB (requires code implementation)
# Then test again
python cart_performance_test.py --host http://$PUBLIC_IP:8080 --output dynamodb_results.json

# Compare results
python compare_databases.py postgres_results.json dynamodb_results.json
```

---

## 📝 Access Pattern Requirements Met

| Requirement | PostgreSQL | DynamoDB | Status |
|-------------|-----------|----------|--------|
| Get cart by ID (<50ms) | 42ms ✅ | Predicted 10ms ✅ | Both meet requirement |
| Create cart | 94ms ⚠️ | Predicted 20ms ✅ | DynamoDB should win |
| Add/update items | Failed ❌ | Predicted 20ms ✅ | DynamoDB advantage |
| Customer history | Fast ✅ | GSI query ✅ | Both supported |
| Handle 50 items | Yes ✅ | Yes ✅ (400KB limit) | Both support |
| 100+ concurrent | Yes ✅ | Unlimited ✅ | Both support |

---

## 💰 Cost Comparison (Estimated)

### For 150 Operations (1 test run)

**DynamoDB:**
- 50 writes (create) × $1.25/1M = $0.0000625
- 50 writes (add items) × $1.25/1M = $0.0000625
- 50 reads (get cart) × $0.25/1M = $0.0000125
- **Total per test:** ~$0.000138

**PostgreSQL (RDS db.t3.micro):**
- Instance cost: $0.017/hour
- For 9-second test: ~$0.000043
- **But:** Runs 24/7 = $12.24/month

**Verdict:**
- **DynamoDB wins for low/variable traffic**
- **PostgreSQL wins for consistent high traffic** (>100K operations/month)

---

## 🔍 Monitoring DynamoDB

### AWS Console

**Navigate to:**
- AWS Console → DynamoDB → Tables → CS6650L2-shopping-carts

**Key Metrics:**
- Read/Write Capacity (should be On-Demand)
- Item Count
- Table Size
- GSI metrics

### CloudWatch Metrics

```powershell
# View metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/DynamoDB \
  --metric-name ConsumedReadCapacityUnits \
  --dimensions Name=TableName,Value=CS6650L2-shopping-carts \
  --start-time 2025-01-19T00:00:00Z \
  --end-time 2025-01-19T23:59:59Z \
  --period 3600 \
  --statistics Sum
```

---

## 🎓 Learning Outcomes

### NoSQL Access Pattern Analysis

✅ Identified simple key-value access patterns  
✅ Recognized no need for complex joins  
✅ Understood write-heavy workload characteristics  
✅ Analyzed natural data expiration needs

### Design Constraints Applied

✅ Even distribution via UUID partition keys  
✅ Cost optimization via embedded items  
✅ Scalability through partition-per-cart design  
✅ Consistency trade-offs (eventual vs strong)

### Implementation Insights

✅ When NoSQL fits: Simple access, high scale, variable load  
✅ When RDBMS fits: Complex queries, transactions, consistency  
✅ Cost models differ significantly (per-request vs per-hour)  
✅ Performance characteristics vary by use case

---

## 📚 Documentation Structure

```
terraform/modules/dynamodb/
├── main.tf                    # DynamoDB table definition
├── variables.tf               # Module inputs
├── outputs.tf                 # Module outputs
└── DESIGN_DECISIONS.md        # 15-page design doc

src/cart/
├── store.go                   # PostgreSQL implementation
└── store_dynamodb.go          # DynamoDB implementation (TODO)

testing/
├── cart_performance_test.py   # Performance test script
└── compare_databases.py       # Comparison tool (TODO)

DYNAMODB_IMPLEMENTATION_GUIDE.md  # This file
```

---

## 🚧 Current Status

### ✅ Completed

- [x] Terraform DynamoDB module
- [x] Design decisions documented
- [x] Infrastructure integration
- [x] Environment variables configured
- [x] AWS deployment ready

### ⏳ Pending (For Full Implementation)

- [ ] Go DynamoDB store implementation
- [ ] Router updates for database switching
- [ ] DynamoDB-specific tests
- [ ] Performance comparison script
- [ ] Results analysis and documentation

---

## 🎯 Next Steps

1. **Deploy current infrastructure:**
   ```powershell
   cd terraform
   terraform apply -auto-approve
   ```

2. **Verify DynamoDB table:**
   ```powershell
   terraform output dynamodb_table_name
   aws dynamodb describe-table --table-name CS6650L2-shopping-carts
   ```

3. **Continue with Go implementation** (if required)

4. **Run performance comparison** (after implementation complete)

---

## 📖 References

- **Design Decisions:** `terraform/modules/dynamodb/DESIGN_DECISIONS.md`
- **PostgreSQL Schema:** `src/cart/SCHEMA_DESIGN.md`
- **Performance Requirements:** `DATABASE_REQUIREMENTS_FULFILLMENT.md`
- **AWS DynamoDB Best Practices:** https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html

---

## ✨ Summary

You now have:
- ✅ Complete DynamoDB infrastructure as code
- ✅ Comprehensive design documentation (15+ pages)
- ✅ Fair comparison framework
- ✅ Ready-to-deploy Terraform modules
- ✅ Clear next steps for implementation

**To deploy:** Run `terraform apply` from the terraform directory!

