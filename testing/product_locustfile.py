from locust import FastHttpUser, task, LoadTestShape
import random


class ProductsUser(FastHttpUser):
    host = "http://13.220.109.58"

    # Known existing product IDs based on seed data in src/product/store.go
    existing_ids = ["1"]

    # Sample product details for POST updates
    sample_products = [
        {"name": "Widget Alpha", "description": "First gen widget", "price": 12.49},
        {"name": "Widget Beta", "description": "Refined widget", "price": 15.99},
        {"name": "Gadget Pro", "description": "Pro gadget", "price": 29.95},
        {"name": "Gizmo Lite", "description": "Lightweight gizmo", "price": 9.99},
        {"name": "Device X", "description": "Experimental device", "price": 22.50},
    ]

    @task(3)
    def get_product_by_id(self):
        """GET /products/{productId}"""
        product_id = random.choice(self.existing_ids)
        self.client.get(f"/products/{product_id}")

    @task(1)
    def post_product_details(self):
        """POST /products/{productId}/details"""
        product_id = random.choice(self.existing_ids)
        body = random.choice(self.sample_products)
        self.client.post(
            f"/products/{product_id}/details",
            json=body,
            headers={"Content-Type": "application/json"},
        )

    @task(2)
    def product_journey(self):
        """Create → Update details → Get → Negative price validation (expect 400)."""
        # 1) Create a new product
        create_body = random.choice(self.sample_products)
        new_id = None
        with self.client.post(
            "/products",
            json=create_body,
            headers={"Content-Type": "application/json"},
            name="POST /products",
            catch_response=True,
        ) as resp:
            if resp.status_code == 201:
                try:
                    new_id = resp.json().get("id")
                    resp.success()
                except Exception as exc:
                    resp.failure(f"invalid JSON response: {exc}")
            else:
                resp.failure(f"unexpected status {resp.status_code}")

        if not new_id:
            return

        # Track created id for subsequent GETs in this user session
        self.existing_ids.append(str(new_id))

        # 2) Update details
        update_body = random.choice(self.sample_products)
        with self.client.post(
            f"/products/{new_id}/details",
            json=update_body,
            headers={"Content-Type": "application/json"},
            name="POST /products/{id}/details",
            catch_response=True,
        ) as resp:
            if resp.status_code == 204:
                resp.success()
            else:
                resp.failure(f"unexpected status {resp.status_code}")

        # 3) Get
        self.client.get(f"/products/{new_id}", name="GET /products/{id}")

        # 4) Negative price (expect 400)
        invalid = {"name": "Bad Price", "description": "fail case", "price": -1}
        with self.client.post(
            f"/products/{new_id}/details",
            json=invalid,
            headers={"Content-Type": "application/json"},
            name="POST /products/{id}/details (neg price)",
            catch_response=True,
        ) as resp:
            if resp.status_code == 400:
                resp.success()
            else:
                resp.failure(f"expected 400, got {resp.status_code}")


class RampingShape(LoadTestShape):
    """Simple ramp: warm-up → ramp-up → sustain → ramp-down."""
    stages = [
        {"duration": 30, "users": 10, "spawn_rate": 5},
        {"duration": 60, "users": 50, "spawn_rate": 10},
        {"duration": 60, "users": 100, "spawn_rate": 20},
        {"duration": 30, "users": 0, "spawn_rate": 10},
    ]

    def tick(self):
        run_time = self.get_run_time()
        for stage in self.stages:
            if run_time < stage["duration"]:
                return stage["users"], stage["spawn_rate"]
            run_time -= stage["duration"]
        return None
