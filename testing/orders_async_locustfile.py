from locust import FastHttpUser, task
import random


class AsyncOrdersUser(FastHttpUser):
    # Set to your ECS public IP or use --host flag when running locust
    # Example: locust -f orders_async_locustfile.py --host=http://YOUR-IP:8080
    host = "http://18.237.83.153:8080"  # Default for local testing
    
    # No wait time - test maximum throughput
    wait_time = lambda self: 0
    
    # Timeouts (async should respond quickly, so use short timeout)
    connection_timeout = 5.0
    network_timeout = 5.0
    
    def on_start(self):
        """Initialize order counter when user starts"""
        self.order_counter = random.randint(100000, 999999)
    
    @task
    def create_order_async(self):
        """
        Create orders via the asynchronous endpoint.
        Expected response time: < 100ms (no blocking!)
        """
        self.order_counter += 1
        
        # Generate random order data
        order_data = {
            "order_id": f"ASYNC-{self.order_counter}",
            "customer_id": random.randint(1000, 9999),
            "status": "pending",
            "items": [
                {
                    "product_id": f"PROD-{random.randint(1, 100):03d}",
                    "quantity": random.randint(1, 5),
                    "price": round(random.uniform(10.0, 200.0), 2)
                }
            ]
        }
        
        # POST to asynchronous endpoint (expect immediate response)
        with self.client.post(
            "/orders/async",
            json=order_data,
            catch_response=True,
            name="POST /orders/async"
        ) as response:
            if response.status_code == 202:  # Accepted
                try:
                    data = response.json()
                    if data.get("status") == "queued":
                        response.success()
                    else:
                        response.failure(f"Unexpected status: {data.get('status')}")
                except:
                    response.failure("Invalid JSON response")
            else:
                response.failure(f"Expected 202, got {response.status_code}")


if __name__ == "__main__":
    print("""
╔═══════════════════════════════════════════════════════════╗
║     ASYNC ORDERS ENDPOINT - LOAD TEST                     ║
╚═══════════════════════════════════════════════════════════╝

Test the asynchronous order processing endpoint:

1. LIGHT LOAD (10 users):
   locust -f orders_async_locustfile.py \\
       --host=http://YOUR-IP:8080 \\
       --users 10 --spawn-rate 2 --run-time 30s --headless

2. MODERATE LOAD (50 users):
   locust -f orders_async_locustfile.py \\
       --host=http://YOUR-IP:8080 \\
       --users 50 --spawn-rate 10 --run-time 60s --headless

3. FLASH SALE (100 users):
   locust -f orders_async_locustfile.py \\
       --host=http://YOUR-IP:8080 \\
       --users 100 --spawn-rate 20 --run-time 60s --headless

4. EXTREME LOAD (500 users):
   locust -f orders_async_locustfile.py \\
       --host=http://YOUR-IP:8080 \\
       --users 500 --spawn-rate 50 --run-time 120s --headless

Expected Results (Async):
• Response Time: ~50-100ms (immediate!)
• Failure Rate: 0% (no timeouts)
• Throughput: ~500-1000 req/s
• Status Code: 202 Accepted

Compare with Sync Endpoint:
• Sync:  30,000ms, 70% failures, 0.33 req/s
• Async: 50ms, 0% failures, 500 req/s
• Improvement: 600x better!

Monitor queue:
aws sqs get-queue-attributes \\
    --queue-url YOUR-QUEUE-URL \\
    --attribute-names ApproximateNumberOfMessages

Watch processing:
aws logs tail /ecs/CS6650L2 --follow

═══════════════════════════════════════════════════════════
    """)

