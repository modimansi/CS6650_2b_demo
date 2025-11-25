# DynamoDB Shopping Cart Design Decisions

## Overview

This document explains the design decisions for implementing shopping carts using DynamoDB as an alternative to PostgreSQL/MySQL, enabling performance comparison.

---

## Access Pattern Analysis

### Required Access Patterns

| Pattern | Frequency | Latency Requirement | Consistency Requirement |
|---------|-----------|--------------------|-----------------------|
| **Get cart by ID** | Very High | <50ms | Eventual OK |
| **Create new cart** | High | <100ms | Strong preferred |
| **Add/update items** | Very High | <100ms | Eventual OK |
| **Get customer carts** | Low | <500ms | Eventual OK |

### Why Shopping Carts Fit NoSQL

✅ **Simple access patterns**: Primarily key-value lookups by cart ID  
✅ **No complex joins**: Cart items embedded in cart document  
✅ **High write volume**: Frequent cart updates as users shop  
✅ **Session-based data**: Natural expiration after checkout or abandonment  
✅ **Horizontal scalability**: Millions of independent carts  
✅ **Eventual consistency acceptable**: Cart updates don't need immediate consistency

---

## Design Decision #1: Partition Key Strategy

### Choice: `cart_id` (String/UUID)

```
Partition Key: cart_id (String - UUID format)
Sort Key: None (not needed)
```

### Why `cart_id`?

✅ **Even distribution**: UUIDs/random IDs distribute evenly across partitions  
✅ **Natural access pattern**: "Get cart by ID" is the primary operation  
✅ **No hot partitions**: Each cart is independent, no single cart gets more traffic  
✅ **Scalable**: Works with millions of carts without rebalancing

### Alternative Considered: `customer_id` as Partition Key

```
Partition Key: customer_id
Sort Key: cart_id
```

❌ **Rejected because:**
- **Hot partitions**: Active customers would create hot spots
- **Uneven distribution**: Some customers shop frequently, others rarely
- **Poor scalability**: Heavy shoppers would overwhelm single partitions
- **Wasted capacity**: Most customers have 0-1 active carts

### Alternative Considered: Composite Key

```
Partition Key: customer_id
Sort Key: created_at
```

❌ **Rejected because:**
- **Primary access pattern mismatch**: We retrieve by cart_id, not customer_id
- **Additional query cost**: Would need GSI to retrieve by cart_id
- **Complexity without benefit**: Adds overhead for rare "customer history" queries

---

## Design Decision #2: Sort Key Need

### Choice: No Sort Key

**Rationale:**
- Each cart is a single item in DynamoDB
- No need to query multiple items under same partition
- Simple key-value access is sufficient
- Items are embedded in cart document, not separate rows

### When a Sort Key Would Be Needed:

If we stored each cart item as a separate DynamoDB item:
```
Partition Key: cart_id
Sort Key: product_id
```

❌ **Rejected this approach because:**
- Multiple read operations to get full cart
- Higher read costs (multiple items vs 1 item)
- More complex to maintain consistency
- Worse performance for "get cart" operation

---

## Design Decision #3: Table Structure

### Choice: Single Table with Embedded Items

```json
{
  "cart_id": "uuid-12345",
  "customer_id": 100,
  "items": [
    {
      "product_id": 1,
      "quantity": 2,
      "product_name": "Widget",
      "product_price": 29.99
    }
  ],
  "created_at": "2025-01-19T10:00:00Z",
  "updated_at": "2025-01-19T10:15:00Z",
  "expiration_time": 1737792000
}
```

### Why Single Table?

✅ **Atomic operations**: Get entire cart in one read  
✅ **Cost efficient**: 1 read unit vs N read units for N items  
✅ **Simple transactions**: Update cart and items together  
✅ **Fast retrieval**: <50ms requirement easily met  
✅ **Natural fit**: Shopping cart is a document/aggregate

### Alternative: Separate Tables

```
Table 1: shopping_carts (cart metadata)
Table 2: cart_items (individual items)
```

❌ **Rejected because:**
- Requires multiple read operations (slower, higher cost)
- Need transactions or risk inconsistency
- Mimics relational model (defeats purpose of NoSQL comparison)
- Poor performance for primary use case

### Item Size Consideration

**Concern**: DynamoDB item size limit is 400 KB

**Analysis**:
- Each item: ~100-200 bytes (product_id, quantity, name, price)
- 400 KB limit = ~2000-4000 items per cart
- Requirement: Support 50 items per cart
- **Verdict**: ✅ Safe with large margin

**If items exceeded limit**: Could use overflow table for carts with 100+ items

---

## Design Decision #4: Attribute Design

### Embedded Items Strategy

```json
{
  "cart_id": "string (PK)",
  "customer_id": "number",
  "items": [
    {
      "product_id": "number",
      "quantity": "number",
      "product_name": "string",      // Denormalized
      "product_price": "number"      // Denormalized
    }
  ],
  "item_count": "number",            // Calculated
  "total_amount": "number",          // Calculated
  "created_at": "string (ISO8601)",
  "updated_at": "string (ISO8601)",
  "expiration_time": "number (Unix timestamp)"  // For TTL
}
```

### Why Embed Product Data?

✅ **Performance**: No need to join with product table  
✅ **Snapshot of purchase**: Price at time of cart creation  
✅ **Self-contained**: Cart has all data needed for display  
✅ **Consistency**: Product price changes don't affect existing carts

❌ **Trade-off**: If product is updated, cart shows old data
- **Acceptable**: Cart should show price when added, not current price
- **Solution**: Refresh prices on checkout if needed

### Why Include Calculated Fields?

```json
{
  "item_count": 5,
  "total_amount": 149.95
}
```

✅ **Avoid client-side calculation**: Faster display  
✅ **Consistent totals**: Calculated once  
✅ **Queryable**: Could filter by total_amount if needed

❌ **Trade-off**: Must update on every item change
- **Solution**: Recalculate on write (minimal overhead)

---

## Design Decision #5: Index Strategy

### Global Secondary Index: CustomerIndex

```
GSI Name: CustomerIndex
Partition Key: customer_id
Projection: ALL
```

### Why CustomerIndex?

✅ **Access pattern**: "Get all carts for customer" (customer history)  
✅ **Enables comparison**: MySQL also supports customer queries  
✅ **Low frequency**: Not primary access pattern, but useful  
✅ **ALL projection**: Avoid additional read cost for full cart data

### Why No Other Indexes?

**Considered**: Index on `created_at` for "recent carts"

❌ **Rejected because:**
- Not a required access pattern
- Can scan CustomerIndex if needed (low volume)
- Additional cost without clear benefit
- Requirement is cart-by-ID lookup, not time-based queries

---

## Design Decision #6: Billing Mode

### Choice: PAY_PER_REQUEST (On-Demand)

**Rationale:**
- **Variable load**: Testing creates burst traffic
- **Cost efficient**: Only pay for actual operations
- **No capacity planning**: Auto-scales instantly
- **Assignment friendly**: No need to predict traffic

### Alternative: Provisioned Capacity

❌ **Rejected because:**
- Must predict RCU/WCU (hard for testing)
- Under-provision = throttling
- Over-provision = wasted cost
- Assignment has variable load patterns

**When to use provisioned:**
- Production with predictable traffic
- Consistent 100 RPS+ load
- Cost optimization at scale

---

## Design Decision #7: Consistency Model

### Choice: Eventual Consistency for Reads

```go
GetItem(&dynamodb.GetItemInput{
    TableName: aws.String("shopping-carts"),
    Key: ...,
    ConsistentRead: aws.Bool(false),  // Eventual consistency
})
```

### Why Eventual?

✅ **Cost**: Half the cost of strongly consistent reads  
✅ **Performance**: Lower latency  
✅ **Acceptable**: Cart updates can tolerate slight delay  
✅ **Use case fit**: Shopping cart doesn't need immediate consistency

### When to Use Strong Consistency?

Use `ConsistentRead: true` for:
- Checkout operation (ensure latest cart state)
- Payment processing (critical accuracy)
- Inventory checks (prevent overselling)

**Our implementation**: Eventual for GET, Strong for checkout

---

## Design Decision #8: TTL (Time-To-Live) Strategy

### Choice: 7-Day Expiration

```json
{
  "expiration_time": 1737792000  // Unix timestamp: created_at + 7 days
}
```

### Why TTL?

✅ **Auto-cleanup**: Abandoned carts deleted automatically  
✅ **Cost savings**: No storage cost for old carts  
✅ **No maintenance**: DynamoDB handles deletion  
✅ **Realistic**: Shopping sessions expire naturally

### Expiration Policy:

| Cart State | TTL |
|------------|-----|
| Active shopping | 7 days |
| Checked out | Deleted immediately (or moved to orders) |
| Abandoned | Auto-deleted after 7 days |

### Alternative: No TTL

❌ **Rejected because:**
- Infinite storage growth
- Manual cleanup scripts needed
- Higher costs
- Cluttered with old data

---

## Design Constraints Satisfied

| Constraint | Solution | Status |
|------------|----------|--------|
| **Even Distribution** | cart_id (UUID) partition key | ✅ |
| **Cost Efficiency** | Embedded items, PAY_PER_REQUEST | ✅ |
| **Scalability** | Partition-per-cart, no hot keys | ✅ |
| **Consistency** | Eventual OK, strong for checkout | ✅ |
| **<50ms retrieval** | Single-item read, no joins | ✅ |

---

## Comparison: DynamoDB vs PostgreSQL

| Aspect | DynamoDB | PostgreSQL |
|--------|----------|------------|
| **Schema** | Schemaless (JSON) | Fixed schema (3 tables) |
| **Access** | Key-value | JOIN-based |
| **Scalability** | Horizontal (unlimited) | Vertical (limited) |
| **Consistency** | Eventual/Strong | Strong (ACID) |
| **Cost Model** | Per-request | Per-hour instance |
| **Joins** | Not supported | Native support |
| **Queries** | Limited | Full SQL |
| **Best for** | Simple access patterns | Complex queries |

---

## Expected Performance

### DynamoDB Predictions:

| Operation | Expected Latency | Cost (per 1000) |
|-----------|------------------|-----------------|
| Get cart by ID | 5-15ms | $0.25 (RCU) |
| Create cart | 10-20ms | $1.25 (WCU) |
| Update items | 15-25ms | $1.25 (WCU) |
| Customer history | 20-50ms | $0.25 (GSI query) |

### vs PostgreSQL (Your Results):

| Operation | PostgreSQL Latency |
|-----------|-------------------|
| Get cart | 42.14ms (P50: 37.79ms) |
| Create cart | 94.47ms (P50: 45.92ms) |
| Add items | Failed (no products) |

**Hypothesis**: DynamoDB will be 2-3x faster for simple key-value operations

---

## Testing Strategy

### Performance Test Requirements:

1. **Same operations as PostgreSQL**: 50 creates, 50 adds, 50 gets
2. **Fair comparison**: Same network conditions, same region
3. **Metrics**: P50, P95, P99, average latency
4. **Output**: `dynamodb_test_results.json` (same format as MySQL)

### Comparison Metrics:

- Latency (DynamoDB vs PostgreSQL)
- Cost per 1000 operations
- Scalability limits
- Consistency trade-offs
- Complexity of implementation

---

## Implementation Files

1. **Terraform**: `terraform/modules/dynamodb/`
2. **Go Store**: `src/cart/store_dynamodb.go`
3. **Router**: Updated `src/cart/router.go` to support both
4. **Tests**: `testing/cart_performance_test_dynamodb.py`
5. **Comparison**: `testing/compare_databases.py`

---

## Next Steps

1. ✅ Create Terraform module (this file)
2. ⏳ Implement Go DynamoDB store
3. ⏳ Update router to support DB switching
4. ⏳ Create comparison test script
5. ⏳ Run performance tests
6. ⏳ Document findings

---

## References

- [DynamoDB Best Practices](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)
- [Partition Key Design](https://aws.amazon.com/blogs/database/choosing-the-right-dynamodb-partition-key/)
- [NoSQL Design Patterns](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/bp-general-nosql-design.html)

