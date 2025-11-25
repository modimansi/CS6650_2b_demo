"""
Shopping Cart Performance Test Script - DynamoDB
Identical to PostgreSQL test for fair comparison

Test Specification:
- Run exactly 150 operations: 50 create, 50 add items, 50 get cart
- Complete test sequence within 5 minutes
- Save results to: dynamodb_test_results.json

Usage:
    python dynamodb_performance_test.py --host http://34.210.69.124:8080
    python dynamodb_performance_test.py --host http://localhost:8080
"""

import requests
import json
import time
import argparse
from datetime import datetime, timezone
import sys
from typing import List, Dict, Any
import random


class CartPerformanceTester:
    """Performance tester for shopping cart API"""
    
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip('/')
        self.results: List[Dict[str, Any]] = []
        self.cart_ids: List[str] = []  # Can be int or UUID string
        
    def log(self, message: str):
        """Print log message with timestamp"""
        timestamp = datetime.now().strftime("%H:%M:%S")
        print(f"[{timestamp}] {message}")
    
    def test_create_cart(self, customer_id: int) -> Dict[str, Any]:
        """
        Test POST /shopping-carts (create cart)
        Returns: Result dictionary with performance metrics
        """
        url = f"{self.base_url}/shopping-carts"
        payload = {"customer_id": customer_id}
        
        start_time = time.time()
        timestamp = datetime.now(timezone.utc).isoformat()
        
        try:
            response = requests.post(
                url,
                json=payload,
                headers={"Content-Type": "application/json"},
                timeout=10
            )
            
            response_time = (time.time() - start_time) * 1000  # Convert to ms
            
            success = response.status_code in [200, 201]
            
            # Extract cart_id if successful (works for both int and UUID)
            if success and response.json():
                cart_id = response.json().get("shopping_cart_id")
                if cart_id:
                    self.cart_ids.append(cart_id)
            
            result = {
                "operation": "create_cart",
                "response_time": round(response_time, 2),
                "success": success,
                "status_code": response.status_code,
                "timestamp": timestamp
            }
            
            return result
            
        except requests.exceptions.RequestException as e:
            response_time = (time.time() - start_time) * 1000
            self.log(f"❌ Error creating cart: {e}")
            
            return {
                "operation": "create_cart",
                "response_time": round(response_time, 2),
                "success": False,
                "status_code": 0,
                "timestamp": timestamp,
                "error": str(e)
            }
    
    def test_add_items(self, cart_id, product_id: int, quantity: int) -> Dict[str, Any]:
        """
        Test POST /shopping-carts/{id}/items (add items)
        Returns: Result dictionary with performance metrics
        """
        url = f"{self.base_url}/shopping-carts/{cart_id}/items"
        payload = {
            "product_id": product_id,
            "quantity": quantity
        }
        
        start_time = time.time()
        timestamp = datetime.now(timezone.utc).isoformat()
        
        try:
            response = requests.post(
                url,
                json=payload,
                headers={"Content-Type": "application/json"},
                timeout=10
            )
            
            response_time = (time.time() - start_time) * 1000  # Convert to ms
            
            success = response.status_code in [200, 204]
            
            result = {
                "operation": "add_items",
                "response_time": round(response_time, 2),
                "success": success,
                "status_code": response.status_code,
                "timestamp": timestamp
            }
            
            return result
            
        except requests.exceptions.RequestException as e:
            response_time = (time.time() - start_time) * 1000
            self.log(f"❌ Error adding items to cart {cart_id}: {e}")
            
            return {
                "operation": "add_items",
                "response_time": round(response_time, 2),
                "success": False,
                "status_code": 0,
                "timestamp": timestamp,
                "error": str(e)
            }
    
    def test_get_cart(self, cart_id) -> Dict[str, Any]:
        """
        Test GET /shopping-carts/{id} (retrieve cart)
        Returns: Result dictionary with performance metrics
        """
        url = f"{self.base_url}/shopping-carts/{cart_id}"
        
        start_time = time.time()
        timestamp = datetime.now(timezone.utc).isoformat()
        
        try:
            response = requests.get(url, timeout=10)
            
            response_time = (time.time() - start_time) * 1000  # Convert to ms
            
            success = response.status_code == 200
            
            result = {
                "operation": "get_cart",
                "response_time": round(response_time, 2),
                "success": success,
                "status_code": response.status_code,
                "timestamp": timestamp
            }
            
            return result
            
        except requests.exceptions.RequestException as e:
            response_time = (time.time() - start_time) * 1000
            self.log(f"❌ Error getting cart {cart_id}: {e}")
            
            return {
                "operation": "get_cart",
                "response_time": round(response_time, 2),
                "success": False,
                "status_code": 0,
                "timestamp": timestamp,
                "error": str(e)
            }
    
    def run_performance_test(self) -> bool:
        """
        Run complete performance test:
        - 50 create cart operations
        - 50 add items operations
        - 50 get cart operations
        
        Returns: True if completed within 5 minutes
        """
        self.log("=" * 70)
        self.log("Shopping Cart Performance Test - DynamoDB")
        self.log("=" * 70)
        self.log(f"Target: {self.base_url}")
        self.log("Test Plan: 50 creates + 50 add items + 50 gets = 150 operations")
        self.log("Time Limit: 5 minutes")
        self.log("=" * 70)
        
        test_start_time = time.time()
        
        # Phase 1: Create 50 shopping carts
        self.log("\n📝 Phase 1: Creating 50 shopping carts...")
        for i in range(1, 51):
            customer_id = random.randint(1, 10000)
            result = self.test_create_cart(customer_id)
            self.results.append(result)
            
            if i % 10 == 0:
                self.log(f"   Created {i}/50 carts...")
        
        successful_creates = sum(1 for r in self.results if r["operation"] == "create_cart" and r["success"])
        self.log(f"✅ Phase 1 Complete: {successful_creates}/50 carts created")
        
        if not self.cart_ids:
            self.log("❌ FATAL: No carts were created successfully. Cannot proceed.")
            return False
        
        # Phase 2: Add items to 50 carts (reuse created carts)
        self.log(f"\n📦 Phase 2: Adding items to 50 carts...")
        for i in range(1, 51):
            # Round-robin through created carts
            cart_id = self.cart_ids[i % len(self.cart_ids)]
            product_id = random.randint(1, 1000)
            quantity = random.randint(1, 5)
            
            result = self.test_add_items(cart_id, product_id, quantity)
            self.results.append(result)
            
            if i % 10 == 0:
                self.log(f"   Added items to {i}/50 carts...")
        
        successful_adds = sum(1 for r in self.results if r["operation"] == "add_items" and r["success"])
        self.log(f"✅ Phase 2 Complete: {successful_adds}/50 add operations successful")
        
        # Phase 3: Retrieve 50 carts
        self.log(f"\n🔍 Phase 3: Retrieving 50 carts...")
        for i in range(1, 51):
            # Round-robin through created carts
            cart_id = self.cart_ids[i % len(self.cart_ids)]
            
            result = self.test_get_cart(cart_id)
            self.results.append(result)
            
            if i % 10 == 0:
                self.log(f"   Retrieved {i}/50 carts...")
        
        successful_gets = sum(1 for r in self.results if r["operation"] == "get_cart" and r["success"])
        self.log(f"✅ Phase 3 Complete: {successful_gets}/50 get operations successful")
        
        # Calculate total time
        total_time = time.time() - test_start_time
        within_time_limit = total_time <= 300  # 5 minutes = 300 seconds
        
        self.log("\n" + "=" * 70)
        self.log("TEST SUMMARY")
        self.log("=" * 70)
        self.log(f"Total Operations: {len(self.results)}/150")
        self.log(f"Total Time: {total_time:.2f} seconds ({total_time/60:.2f} minutes)")
        self.log(f"Time Limit: {'✅ PASS' if within_time_limit else '❌ FAIL'} (must be < 5 minutes)")
        self.log("")
        self.log(f"Create Cart:  {successful_creates}/50 successful")
        self.log(f"Add Items:    {successful_adds}/50 successful")
        self.log(f"Get Cart:     {successful_gets}/50 successful")
        self.log(f"Total Success: {successful_creates + successful_adds + successful_gets}/150")
        
        # Calculate performance metrics
        self._print_performance_metrics()
        
        return within_time_limit
    
    def _print_performance_metrics(self):
        """Print detailed performance metrics"""
        if not self.results:
            return
        
        self.log("\n" + "=" * 70)
        self.log("PERFORMANCE METRICS")
        self.log("=" * 70)
        
        for operation in ["create_cart", "add_items", "get_cart"]:
            op_results = [r for r in self.results if r["operation"] == operation and r["success"]]
            
            if op_results:
                response_times = [r["response_time"] for r in op_results]
                avg_time = sum(response_times) / len(response_times)
                min_time = min(response_times)
                max_time = max(response_times)
                
                # Calculate percentiles
                sorted_times = sorted(response_times)
                p50 = sorted_times[len(sorted_times) // 2]
                p95 = sorted_times[int(len(sorted_times) * 0.95)]
                p99 = sorted_times[int(len(sorted_times) * 0.99)]
                
                self.log(f"\n{operation}:")
                self.log(f"  Average: {avg_time:.2f} ms")
                self.log(f"  Min:     {min_time:.2f} ms")
                self.log(f"  Max:     {max_time:.2f} ms")
                self.log(f"  P50:     {p50:.2f} ms")
                self.log(f"  P95:     {p95:.2f} ms")
                self.log(f"  P99:     {p99:.2f} ms")
                
                # Check if meets <50ms requirement
                if operation == "get_cart":
                    meets_requirement = avg_time < 50
                    self.log(f"  <50ms requirement: {'✅ PASS' if meets_requirement else '❌ FAIL'}")
    
    def save_results(self, filename: str = "dynamodb_test_results.json"):
        """Save test results to JSON file"""
        try:
            with open(filename, 'w') as f:
                json.dump(self.results, f, indent=2)
            
            self.log(f"\n💾 Results saved to: {filename}")
            self.log(f"   Total records: {len(self.results)}")
            return True
            
        except Exception as e:
            self.log(f"❌ Error saving results: {e}")
            return False


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(
        description="Shopping Cart Performance Test - DynamoDB"
    )
    parser.add_argument(
        '--host',
        type=str,
        default='http://localhost:8080',
        help='Base URL of the API (default: http://localhost:8080)'
    )
    parser.add_argument(
        '--output',
        type=str,
        default='dynamodb_test_results.json',
        help='Output filename (default: dynamodb_test_results.json)'
    )
    
    args = parser.parse_args()
    
    # Create tester instance
    tester = CartPerformanceTester(args.host)
    
    # Run test
    success = tester.run_performance_test()
    
    # Save results
    tester.save_results(args.output)
    
    # Exit with appropriate code
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()


