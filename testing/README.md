# Load Testing Guide

## Available Test Files

| File | Endpoint | Wait Time | Purpose |
|------|----------|-----------|---------|
| `product_locustfile.py` | `/products` | 0.01s | Product search testing |
| `orders_locustfile.py` | `/orders/sync` | 0s (none) | Orders sync endpoint stress test |

## Quick Start

### Test Orders Endpoint (Local)

```bash
cd testing
locust -f orders_locustfile.py --host=http://localhost:8080
```

Then open: **http://localhost:8089** in your browser

### Test Orders Endpoint (AWS ECS)

First, get your ECS public IP:
```bash
cd terraform
./get_public_ip.sh  # or get_public_ip.ps1 for Windows
```

Then run Locust:
```bash
cd testing
locust -f orders_locustfile.py --host=http://YOUR-PUBLIC-IP:8080
```

## Usage Examples

### 1. Interactive Web UI Mode (Recommended)

```bash
locust -f orders_locustfile.py --host=http://localhost:8080
```

- Open http://localhost:8089
- Enter number of users (e.g., 10)
- Enter spawn rate (e.g., 2 users/second)
- Click "Start Swarming"

### 2. Headless Mode (Command Line)

```bash
# Light load: 10 users for 60 seconds
locust -f orders_locustfile.py \
    --host=http://localhost:8080 \
    --users 10 \
    --spawn-rate 2 \
    --run-time 60s \
    --headless

# Moderate load: 30 users for 2 minutes
locust -f orders_locustfile.py \
    --host=http://localhost:8080 \
    --users 30 \
    --spawn-rate 5 \
    --run-time 120s \
    --headless

# Heavy load: 100 users for 5 minutes
locust -f orders_locustfile.py \
    --host=http://localhost:8080 \
    --users 100 \
    --spawn-rate 10 \
    --run-time 300s \
    --headless
```

### 3. Generate HTML Report

```bash
locust -f orders_locustfile.py \
    --host=http://localhost:8080 \
    --users 30 \
    --spawn-rate 5 \
    --run-time 180s \
    --headless \
    --html report.html
```

## Expected Behavior

### Orders Endpoint (Synchronous)

The `/orders/sync` endpoint has a **3-second blocking delay** for payment processing.

**With NO wait time between requests:**
- Each user spawns new requests immediately after the previous one completes
- This creates **continuous load** on the system
- Tests connection pool limits and concurrent request handling

**Expected Results:**

| Users | Throughput | Avg Response Time | Notes |
|-------|------------|-------------------|-------|
| 10 | ~3.3 req/s | ~3000ms | Stable |
| 30 | ~10 req/s | ~3000-3500ms | Moderate load |
| 50 | ~15 req/s | ~3500-5000ms | High load |
| 100 | ~20 req/s | >5000ms | Connection issues likely |

**Failure Indicators:**
- ❌ Response times > 5 seconds
- ❌ Connection errors
- ❌ Timeout errors
- ❌ 5xx server errors

## Testing Strategy

### 1. Baseline Test
**Goal:** Verify functionality  
**Config:** 5 users, 1 min

```bash
locust -f orders_locustfile.py --host=http://localhost:8080 \
    --users 5 --spawn-rate 1 --run-time 60s --headless
```

### 2. Load Test
**Goal:** Normal production load  
**Config:** 20 users, 3 min

```bash
locust -f orders_locustfile.py --host=http://localhost:8080 \
    --users 20 --spawn-rate 2 --run-time 180s --headless
```

### 3. Stress Test
**Goal:** Find breaking point  
**Config:** 100 users, 5 min

```bash
locust -f orders_locustfile.py --host=http://localhost:8080 \
    --users 100 --spawn-rate 10 --run-time 300s --headless
```

## Understanding Results

### Sample Output

```
Type     Name              # reqs  # fails  Avg    Min    Max   Median  req/s
------------------------------------------------------------------------------
POST     /orders/sync       1000      0    3001   2998   3456    3000   3.3
------------------------------------------------------------------------------
                           1000      0    3001   2998   3456    3000   3.3

Response time percentiles (approximated)
Type     Name                50%    66%    75%    80%    90%    95%    98%    99%  99.9% 100%
-------------------------------------------------------------------------------------------
POST     /orders/sync       3000   3010   3020   3025   3050   3100   3200   3300   3450 3456
```

### Key Metrics

| Metric | Description | Good | Warning | Critical |
|--------|-------------|------|---------|----------|
| **Avg Response Time** | Average time per request | ~3000ms | 3000-4000ms | >5000ms |
| **Failure Rate** | % of failed requests | 0% | 1-5% | >5% |
| **Requests/sec** | Throughput | 3-10 | 10-20 | >20 |
| **P95 Response Time** | 95th percentile | <4000ms | 4000-6000ms | >6000ms |

## Monitoring

### Watch Server Logs (AWS)

```bash
aws logs tail /ecs/CS6650L2 --follow
```

### Check ECS Task Status

```bash
aws ecs describe-services \
    --cluster CS6650L2-cluster \
    --services CS6650L2
```

### Check Task CPU/Memory

View in AWS Console:
1. Go to ECS Console
2. Click on CS6650L2-cluster
3. Click on CS6650L2 service
4. View "Metrics" tab

## Troubleshooting

### Connection Refused
**Problem:** Can't connect to server  
**Solution:** Check server is running and accessible
```bash
curl http://localhost:8080/health
```

### High Failure Rate (>10%)
**Problem:** Many requests failing  
**Cause:** Too many concurrent users  
**Solution:** Reduce user count or increase spawn rate gradually

### Timeout Errors
**Problem:** Requests timing out  
**Cause:** Server overloaded or network issues  
**Solution:** 
- Reduce concurrent users
- Check server resources
- Monitor logs for errors

### Connection Pool Exhausted
**Problem:** "Connection pool limit" errors  
**Cause:** Too many concurrent requests (synchronous blocking)  
**Solution:** 
- This is expected with synchronous processing
- Reduce user count
- This demonstrates why async processing is needed!

## Tips

1. **Start Small:** Begin with 5-10 users
2. **Ramp Gradually:** Use slow spawn rate (1-2 users/sec)
3. **Monitor Logs:** Watch server logs during test
4. **Save Reports:** Generate HTML reports for documentation
5. **Test Locally First:** Validate before testing AWS
6. **No Wait Time = Heavy Load:** With 0 wait time, system is under constant pressure

## Comparing Sync vs Async (Phase 2)

After implementing async endpoint, compare performance:

```bash
# Test sync endpoint
locust -f orders_locustfile.py --host=http://localhost:8080 \
    --users 30 --spawn-rate 5 --run-time 180s --headless --html sync_report.html

# Test async endpoint (future)
locust -f orders_async_locustfile.py --host=http://localhost:8080 \
    --users 30 --spawn-rate 5 --run-time 180s --headless --html async_report.html
```

**Expected Difference:**
- Sync: ~3000ms avg, ~10 req/s
- Async: ~50ms avg, ~600 req/s
- **60x improvement!**

## Clean Up After Testing

Stop Locust: `Ctrl + C`

If you're done with AWS resources:
```bash
cd terraform
terraform destroy
```

---

Happy load testing! 🚀

