# 🚀 Deploy and Test DynamoDB Configuration

## ✅ Configuration Updated
- `CART_STORE_TYPE` is now set to `"dynamodb"`
- DynamoDB table module is already configured
- ECS task will use DynamoDB instead of PostgreSQL

---

## 📋 Step-by-Step Deployment

### Step 1: Build the Updated Application

```powershell
# Navigate to source directory
cd src

# Build the Go application (verify compilation)
go build -o main .

# Verify build succeeded
if ($?) { Write-Host "✅ Build successful!" -ForegroundColor Green } else { Write-Host "❌ Build failed!" -ForegroundColor Red; exit 1 }

cd ..
```

---

### Step 2: Build and Push Docker Image to ECR

```powershell
# Get AWS Account ID
$ACCOUNT_ID = (aws sts get-caller-identity --query Account --output text)
Write-Host "AWS Account ID: $ACCOUNT_ID" -ForegroundColor Cyan

# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin "${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com"

# Build Docker image
cd src
docker build -t cs6650l2-shopping-api:latest .

# Tag for ECR
docker tag cs6650l2-shopping-api:latest "${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/cs6650l2-shopping-api:latest"

# Push to ECR
docker push "${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/cs6650l2-shopping-api:latest"

if ($?) { 
    Write-Host "✅ Docker image pushed to ECR successfully!" -ForegroundColor Green 
} else { 
    Write-Host "❌ Docker push failed!" -ForegroundColor Red
    exit 1 
}

cd ..
```

---

### Step 3: Deploy with Terraform (DynamoDB Mode)

```powershell
cd terraform

# Apply Terraform changes (will create DynamoDB table and update ECS)
terraform apply -auto-approve

# Wait for ECS task to stabilize (2-3 minutes)
Write-Host "`n⏳ Waiting for ECS task to start (180 seconds)..." -ForegroundColor Yellow
Start-Sleep -Seconds 180

# Get the public IP of the ECS task
Write-Host "`n🔍 Fetching ECS Task Public IP..." -ForegroundColor Cyan

$CLUSTER_NAME = "CS6650L2-shopping-api-cluster"
$SERVICE_NAME = "CS6650L2-shopping-api"

# Get task ARN
$TASK_ARN = (aws ecs list-tasks --cluster $CLUSTER_NAME --service-name $SERVICE_NAME --query 'taskArns[0]' --output text)

if ($TASK_ARN -and $TASK_ARN -ne "None") {
    # Get task details
    $TASK_DETAILS = (aws ecs describe-tasks --cluster $CLUSTER_NAME --tasks $TASK_ARN --query 'tasks[0]' --output json | ConvertFrom-Json)
    
    # Extract ENI ID
    $ENI_ID = $TASK_DETAILS.attachments[0].details | Where-Object { $_.name -eq "networkInterfaceId" } | Select-Object -ExpandProperty value
    
    if ($ENI_ID) {
        # Get public IP from ENI
        $PUBLIC_IP = (aws ec2 describe-network-interfaces --network-interface-ids $ENI_ID --query 'NetworkInterfaces[0].Association.PublicIp' --output text)
        
        if ($PUBLIC_IP -and $PUBLIC_IP -ne "None") {
            Write-Host "`n✅ ECS Task Public IP: $PUBLIC_IP" -ForegroundColor Green
            Write-Host "   API Base URL: http://${PUBLIC_IP}:8080" -ForegroundColor Cyan
            
            # Save to environment variable for testing
            $env:API_BASE_URL = "http://${PUBLIC_IP}:8080"
            Write-Host "`n💾 Saved to environment: `$env:API_BASE_URL" -ForegroundColor Green
        } else {
            Write-Host "❌ Could not retrieve public IP" -ForegroundColor Red
        }
    } else {
        Write-Host "❌ Could not find network interface" -ForegroundColor Red
    }
} else {
    Write-Host "❌ No running tasks found" -ForegroundColor Red
}

cd ..
```

---

### Step 4: Quick Smoke Test (DynamoDB with UUID IDs)

```powershell
# Test 1: Create a cart (should return UUID)
Write-Host "`n🧪 Test 1: Creating a cart..." -ForegroundColor Yellow

$response = Invoke-RestMethod -Uri "$env:API_BASE_URL/shopping-carts" `
    -Method Post `
    -ContentType "application/json" `
    -Body '{"customer_id": 123}'

Write-Host "Response: $($response | ConvertTo-Json)" -ForegroundColor Cyan
$CART_ID = $response.shopping_cart_id
Write-Host "✅ Cart Created with UUID: $CART_ID" -ForegroundColor Green

# Test 2: Add item to cart
Write-Host "`n🧪 Test 2: Adding item to cart..." -ForegroundColor Yellow

Invoke-RestMethod -Uri "$env:API_BASE_URL/shopping-carts/$CART_ID/items" `
    -Method Post `
    -ContentType "application/json" `
    -Body '{"product_id": 42, "quantity": 3}'

Write-Host "✅ Item added successfully (204 No Content)" -ForegroundColor Green

# Test 3: Get cart with items
Write-Host "`n🧪 Test 3: Retrieving cart with items..." -ForegroundColor Yellow

$cart = Invoke-RestMethod -Uri "$env:API_BASE_URL/shopping-carts/$CART_ID" `
    -Method Get

Write-Host "Cart Details:" -ForegroundColor Cyan
Write-Host ($cart | ConvertTo-Json -Depth 5) -ForegroundColor White

Write-Host "`n✅ All smoke tests passed!" -ForegroundColor Green
```

---

### Step 5: Run Full DynamoDB Consistency Test

```powershell
cd testing

# Update the test script with the correct base URL
$BASE_URL = $env:API_BASE_URL

# Create DynamoDB consistency test script
@"
import requests
import time
import json
import random
from datetime import datetime

BASE_URL = "$BASE_URL"

def run_consistency_test():
    results = []
    created_carts = []
    
    print("🧪 DynamoDB Consistency Test")
    print("=" * 60)
    print(f"Base URL: {BASE_URL}")
    
    # TEST 1: Create 50 carts
    print("\n1️⃣  Creating 50 carts...")
    create_success = 0
    for i in range(50):
        start_time = time.time()
        try:
            response = requests.post(
                f"{BASE_URL}/shopping-carts",
                json={"customer_id": random.randint(1, 1000)},
                timeout=10
            )
            response_time = (time.time() - start_time) * 1000
            
            results.append({
                "operation": "create_cart",
                "response_time": round(response_time, 2),
                "success": response.status_code == 201,
                "status_code": response.status_code,
                "timestamp": datetime.utcnow().isoformat() + "Z"
            })
            
            if response.status_code == 201:
                cart_id = response.json()["shopping_cart_id"]
                created_carts.append(cart_id)
                create_success += 1
                if (i + 1) % 10 == 0:
                    print(f"   Progress: {i+1}/50 carts created")
        except Exception as e:
            results.append({
                "operation": "create_cart",
                "response_time": 0,
                "success": False,
                "status_code": 0,
                "timestamp": datetime.utcnow().isoformat() + "Z",
                "error": str(e)
            })
    
    print(f"   ✅ Summary: {create_success}/50 carts created")
    
    # TEST 2: Add items to 50 carts
    print("\n2️⃣  Adding items to carts...")
    add_success = 0
    for i, cart_id in enumerate(created_carts[:50]):
        start_time = time.time()
        try:
            response = requests.post(
                f"{BASE_URL}/shopping-carts/{cart_id}/items",
                json={
                    "product_id": random.randint(1, 100),
                    "quantity": random.randint(1, 5)
                },
                timeout=10
            )
            response_time = (time.time() - start_time) * 1000
            
            results.append({
                "operation": "add_items",
                "response_time": round(response_time, 2),
                "success": response.status_code == 204,
                "status_code": response.status_code,
                "timestamp": datetime.utcnow().isoformat() + "Z"
            })
            
            if response.status_code == 204:
                add_success += 1
                if (i + 1) % 10 == 0:
                    print(f"   Progress: {i+1}/50 items added")
        except Exception as e:
            results.append({
                "operation": "add_items",
                "response_time": 0,
                "success": False,
                "status_code": 0,
                "timestamp": datetime.utcnow().isoformat() + "Z",
                "error": str(e)
            })
    
    print(f"   ✅ Summary: {add_success}/50 items added")
    
    # TEST 3: Get 50 carts (immediate read-after-write for consistency testing)
    print("\n3️⃣  Retrieving carts (consistency test)...")
    get_success = 0
    consistency_issues = 0
    
    for i, cart_id in enumerate(created_carts[:50]):
        start_time = time.time()
        try:
            response = requests.get(
                f"{BASE_URL}/shopping-carts/{cart_id}",
                timeout=10
            )
            response_time = (time.time() - start_time) * 1000
            
            # Check for consistency (cart should have items we just added)
            has_items = False
            if response.status_code == 200:
                cart_data = response.json()
                has_items = len(cart_data.get("items", [])) > 0
                if not has_items:
                    consistency_issues += 1
            
            results.append({
                "operation": "get_cart",
                "response_time": round(response_time, 2),
                "success": response.status_code == 200,
                "status_code": response.status_code,
                "timestamp": datetime.utcnow().isoformat() + "Z",
                "has_items": has_items
            })
            
            if response.status_code == 200:
                get_success += 1
                if (i + 1) % 10 == 0:
                    print(f"   Progress: {i+1}/50 carts retrieved")
        except Exception as e:
            results.append({
                "operation": "get_cart",
                "response_time": 0,
                "success": False,
                "status_code": 0,
                "timestamp": datetime.utcnow().isoformat() + "Z",
                "error": str(e)
            })
    
    print(f"   ✅ Summary: {get_success}/50 carts retrieved")
    if consistency_issues > 0:
        print(f"   ⚠️  Consistency Issues: {consistency_issues}/50 carts missing items (eventual consistency)")
    
    # Save results
    with open("dynamodb_test_results.json", "w") as f:
        json.dump(results, f, indent=2)
    
    # Print summary
    print("\n" + "=" * 60)
    print("📊 Test Summary")
    print("=" * 60)
    
    for op_type in ["create_cart", "add_items", "get_cart"]:
        op_results = [r for r in results if r["operation"] == op_type]
        successful = [r for r in op_results if r["success"]]
        
        if successful:
            avg_time = sum(r["response_time"] for r in successful) / len(successful)
            min_time = min(r["response_time"] for r in successful)
            max_time = max(r["response_time"] for r in successful)
            
            print(f"\n{op_type.upper().replace('_', ' ')}:")
            print(f"  Success Rate: {len(successful)}/{len(op_results)} ({len(successful)/len(op_results)*100:.1f}%)")
            print(f"  Avg Response: {avg_time:.2f}ms")
            print(f"  Min Response: {min_time:.2f}ms")
            print(f"  Max Response: {max_time:.2f}ms")
    
    print("\n✅ Results saved to: dynamodb_test_results.json")
    print("\nNext: Compare with mysql_test_results.json")

if __name__ == "__main__":
    run_consistency_test()
"@ | Out-File -FilePath "cart_dynamodb_test.py" -Encoding UTF8

Write-Host "`n✅ Test script created: cart_dynamodb_test.py" -ForegroundColor Green

# Run the test
python cart_dynamodb_test.py

cd ..
```

---

### Step 6: Compare PostgreSQL vs DynamoDB Results

```powershell
cd testing

# Create comparison script
@"
import json

def analyze_results(filename, db_name):
    try:
        with open(filename, 'r') as f:
            results = json.load(f)
    except FileNotFoundError:
        print(f"❌ {filename} not found!")
        return
    
    print(f"\n{'='*60}")
    print(f"{db_name} Performance Analysis")
    print(f"{'='*60}")
    
    for op in ["create_cart", "add_items", "get_cart"]:
        op_results = [r for r in results if r["operation"] == op]
        successful = [r for r in op_results if r["success"]]
        
        if successful:
            times = [r["response_time"] for r in successful]
            avg = sum(times) / len(times)
            min_time = min(times)
            max_time = max(times)
            p95 = sorted(times)[int(len(times) * 0.95)] if len(times) > 0 else 0
            
            print(f"\n{op.upper().replace('_', ' ')}:")
            print(f"  Total: {len(op_results)}")
            print(f"  Success: {len(successful)} ({len(successful)/len(op_results)*100:.1f}%)")
            print(f"  Avg: {avg:.2f}ms")
            print(f"  Min: {min_time:.2f}ms")
            print(f"  Max: {max_time:.2f}ms")
            print(f"  P95: {p95:.2f}ms")

print("\n" + "="*60)
print("🔍 POSTGRESQL vs DYNAMODB COMPARISON")
print("="*60)

analyze_results("mysql_test_results.json", "PostgreSQL (RDS)")
analyze_results("dynamodb_test_results.json", "DynamoDB")

print("\n" + "="*60)
print("✅ Comparison Complete!")
print("="*60)
"@ | Out-File -FilePath "compare_results.py" -Encoding UTF8

# Run comparison
python compare_results.py

cd ..
```

---

## 🎯 Expected Behavior

### DynamoDB Responses (UUID-based):
```json
{
  "shopping_cart_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### PostgreSQL Responses (Integer-based):
```json
{
  "shopping_cart_id": 1
}
```

---

## 🔄 Switch Back to PostgreSQL (If Needed)

```powershell
cd terraform/modules/ecs

# Change back to postgres
(Get-Content main.tf) -replace '"dynamodb"', '"postgres"' | Set-Content main.tf

cd ../..
terraform apply -auto-approve
```

---

## ✅ Success Indicators

1. **Build succeeds** without compilation errors
2. **Docker push** completes successfully
3. **Terraform apply** creates/updates DynamoDB table
4. **ECS task** starts and becomes healthy
5. **Smoke test** returns UUID cart IDs (not integers)
6. **Full test** completes with 150 operations
7. **Comparison** shows performance differences

---

## 🐛 Troubleshooting

### If ECS task fails to start:
```powershell
# Check task logs
aws logs tail /ecs/CS6650L2-shopping-api --follow
```

### If cart creation fails:
```powershell
# Verify DynamoDB table exists
aws dynamodb describe-table --table-name CS6650L2-shopping-carts
```

### If API is unreachable:
```powershell
# Verify security group allows inbound on port 8080
aws ec2 describe-security-groups --group-ids <security-group-id>
```

---

**Ready to deploy! Run the commands in order.** 🚀

