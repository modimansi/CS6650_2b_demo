from locust import FastHttpUser, task
import random


class OrdersUser(FastHttpUser):
    # Set to your ECS public IP or use --host flag when running locust
    # Example: locust -f orders_locustfile.py --host=http://YOUR-IP:8080
    host = "http://localhost:8080"  # Default for local testing

    # No wait time - continuous load to test synchronous blocking behavior
    wait_time = lambda self: 0
    
    # Set connection and request timeouts (30 seconds)
    # Requests taking longer than this will fail with timeout
    connection_timeout = 30.0
    network_timeout = 30.0

    def on_start(self):
        """Initialize order counter when user starts"""
        self.order_counter = random.randint(100000, 999999)

    @task
    def create_order_sync(self):
        """
        Create orders continuously via the synchronous endpoint.
        Each request will block for ~3 seconds due to payment processing.
        With no wait time, this will create heavy load and test:
        - Connection pool limits
        - Concurrent request handling
        - System behavior under sustained load
        """
        self.order_counter += 1
        
        # Generate random order data
        order_data = {
            "order_id": f"ORD-{self.order_counter}",
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
        
        # POST to synchronous endpoint (expect ~3 second response)
        self.client.post("/orders/sync", json=order_data, name="POST /orders/sync")

