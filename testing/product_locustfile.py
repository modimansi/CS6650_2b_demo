from locust import FastHttpUser, task
import random


COMMON_TERMS = [
    # Name terms (brands are embedded in names)
    "alpha",
    "beta",
    "gamma",
    # Category terms
    "books",
    "electronics",
    "home",
]


class ProductsUser(FastHttpUser):
    # Set this to your target, e.g., http://localhost:8080 for local runs
    host = "http://44.251.132.143:8080"

    # Minimal wait to generate constant pressure
    wait_time = lambda self: 0.01

    @task(5)
    def search_by_name(self):
        term = random.choice(["alpha", "beta", "gamma"])  # common terms
        self.client.get(f"/products?name={term}", name="GET /products?name=")

    @task(3)
    def search_by_category(self):
        term = random.choice(["books", "electronics", "home"])  # common categories
        self.client.get(f"/products?category={term}", name="GET /products?category=")

    @task(2)
    def search_by_both(self):
        name = random.choice(["alpha", "beta"]) 
        cat = random.choice(["books", "electronics"]) 
        self.client.get(f"/products?name={name}&category={cat}", name="GET /products?name&category")

    import logging, random
    logger = logging.getLogger("locust")
