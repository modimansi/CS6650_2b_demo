"""
DynamoDB Consistency Testing Script

Testing Objectives:
- Observe read-after-write consistency behavior
- Measure how quickly consistency is achieved
- Document how eventual consistency affects user experience

Test Scenarios:
1. Create cart then immediately retrieve it
2. Add item then immediately fetch cart items
3. Rapid updates to the same cart from multiple clients

Usage:
    python cart_consistency_test.py --host http://44.242.214.61:8080
"""

import requests
import json
import time
import argparse
from datetime import datetime, timezone
from typing import List, Dict, Any
import random


class ConsistencyTester:
    """Tests DynamoDB eventual consistency behavior"""
    
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip('/')
        self.results: List[Dict[str, Any]] = []
        
    def log(self, message: str):
        """Print log message with timestamp"""
        timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
        print(f"[{timestamp}] {message}")
    
    def test_create_then_read(self) -> Dict[str, Any]:
        """
        Test read-after-write consistency
        Creates a cart, then immediately reads it
        """
        self.log("\n=== Test 1: Create Then Immediate Read ===")
        
        # Create cart
        create_start = time.time()
        response = requests.post(
            f"{self.base_url}/shopping-carts",
            json={"customer_id": random.randint(1, 10000)},
            timeout=10
        )
        create_time = (time.time() - create_start) * 1000
        
        if response.status_code not in [200, 201]:
            self.log(f"❌ Create failed: {response.status_code}")
            return {"test": "create_then_read", "success": False}
        
        cart_id = response.json().get("shopping_cart_id")
        self.log(f"✓ Created cart: {cart_id} ({create_time:.2f}ms)")
        
        # Immediately read cart (eventual consistency test)
        delays = []
        for attempt in range(5):
            read_start = time.time()
            read_response = requests.get(
                f"{self.base_url}/shopping-carts/{cart_id}",
                timeout=10
            )
            read_time = (time.time() - read_start) * 1000
            delay = (time.time() - (create_start + create_time/1000)) * 1000
            
            if read_response.status_code == 200:
                self.log(f"✓ Read #{attempt+1}: Success ({read_time:.2f}ms, delay: {delay:.2f}ms)")
                delays.append(delay)
            elif read_response.status_code == 404:
                self.log(f"⚠ Read #{attempt+1}: Not found yet ({read_time:.2f}ms, delay: {delay:.2f}ms)")
                delays.append(None)
            
            time.sleep(0.01)  # 10ms between attempts
        
        consistent_reads = sum(1 for d in delays if d is not None)
        
        result = {
            "test": "create_then_read",
            "cart_id": cart_id,
            "create_time_ms": create_time,
            "read_attempts": len(delays),
            "consistent_reads": consistent_reads,
            "consistency_rate": consistent_reads / len(delays),
            "delays_ms": delays,
            "timestamp": datetime.now(timezone.utc).isoformat()
        }
        
        self.log(f"Consistency rate: {consistent_reads}/{len(delays)} ({result['consistency_rate']*100:.1f}%)")
        
        return result
    
    def test_write_then_read_items(self) -> Dict[str, Any]:
        """
        Test item update consistency
        Creates cart, adds item, immediately reads cart
        """
        self.log("\n=== Test 2: Add Item Then Immediate Read ===")
        
        # Create cart
        response = requests.post(
            f"{self.base_url}/shopping-carts",
            json={"customer_id": random.randint(1, 10000)},
            timeout=10
        )
        
        if response.status_code not in [200, 201]:
            return {"test": "write_then_read_items", "success": False}
        
        cart_id = response.json().get("shopping_cart_id")
        self.log(f"✓ Created cart: {cart_id}")
        
        # Add item
        product_id = random.randint(1, 1000)
        add_start = time.time()
        add_response = requests.post(
            f"{self.base_url}/shopping-carts/{cart_id}/items",
            json={"product_id": product_id, "quantity": 2},
            timeout=10
        )
        add_time = (time.time() - add_start) * 1000
        
        if add_response.status_code not in [200, 204]:
            self.log(f"❌ Add item failed: {add_response.status_code}")
            return {"test": "write_then_read_items", "success": False}
        
        self.log(f"✓ Added product {product_id} ({add_time:.2f}ms)")
        
        # Immediately read cart items
        item_found_attempts = []
        for attempt in range(5):
            read_start = time.time()
            read_response = requests.get(
                f"{self.base_url}/shopping-carts/{cart_id}",
                timeout=10
            )
            read_time = (time.time() - read_start) * 1000
            delay = (time.time() - (add_start + add_time/1000)) * 1000
            
            if read_response.status_code == 200:
                items = read_response.json().get("items", [])
                found = any(item["product_id"] == product_id for item in items)
                item_found_attempts.append(found)
                
                if found:
                    self.log(f"✓ Read #{attempt+1}: Item visible ({read_time:.2f}ms, delay: {delay:.2f}ms)")
                else:
                    self.log(f"⚠ Read #{attempt+1}: Item not visible yet ({read_time:.2f}ms, delay: {delay:.2f}ms)")
            
            time.sleep(0.01)
        
        consistent_reads = sum(item_found_attempts)
        
        result = {
            "test": "write_then_read_items",
            "cart_id": cart_id,
            "product_id": product_id,
            "add_time_ms": add_time,
            "read_attempts": len(item_found_attempts),
            "item_visible_count": consistent_reads,
            "consistency_rate": consistent_reads / len(item_found_attempts),
            "timestamp": datetime.now(timezone.utc).isoformat()
        }
        
        self.log(f"Item visibility: {consistent_reads}/{len(item_found_attempts)} ({result['consistency_rate']*100:.1f}%)")
        
        return result
    
    def test_rapid_updates(self) -> Dict[str, Any]:
        """
        Test rapid updates to same cart
        Simulates multiple clients updating cart simultaneously
        """
        self.log("\n=== Test 3: Rapid Updates to Same Cart ===")
        
        # Create cart
        response = requests.post(
            f"{self.base_url}/shopping-carts",
            json={"customer_id": random.randint(1, 10000)},
            timeout=10
        )
        
        if response.status_code not in [200, 201]:
            return {"test": "rapid_updates", "success": False}
        
        cart_id = response.json().get("shopping_cart_id")
        self.log(f"✓ Created cart: {cart_id}")
        
        # Perform 5 rapid updates
        updates = []
        for i in range(5):
            product_id = random.randint(1, 100)
            start = time.time()
            response = requests.post(
                f"{self.base_url}/shopping-carts/{cart_id}/items",
                json={"product_id": product_id, "quantity": 1},
                timeout=10
            )
            update_time = (time.time() - start) * 1000
            
            updates.append({
                "product_id": product_id,
                "time_ms": update_time,
                "success": response.status_code in [200, 204]
            })
            
            self.log(f"Update {i+1}: Product {product_id} ({update_time:.2f}ms)")
            
            time.sleep(0.005)  # 5ms between updates
        
        # Read final cart state
        time.sleep(0.05)  # 50ms delay for consistency
        read_response = requests.get(
            f"{self.base_url}/shopping-carts/{cart_id}",
            timeout=10
        )
        
        final_items = []
        if read_response.status_code == 200:
            final_items = read_response.json().get("items", [])
            self.log(f"✓ Final cart has {len(final_items)} items")
        
        result = {
            "test": "rapid_updates",
            "cart_id": cart_id,
            "updates_sent": len(updates),
            "updates_successful": sum(1 for u in updates if u["success"]),
            "final_items_count": len(final_items),
            "updates": updates,
            "timestamp": datetime.now(timezone.utc).isoformat()
        }
        
        return result
    
    def run_all_tests(self, iterations: int = 3):
        """Run all consistency tests multiple times"""
        self.log("=" * 70)
        self.log("DynamoDB Consistency Testing")
        self.log("=" * 70)
        self.log(f"Target: {self.base_url}")
        self.log(f"Iterations: {iterations}")
        self.log("=" * 70)
        
        all_results = {
            "create_then_read": [],
            "write_then_read_items": [],
            "rapid_updates": []
        }
        
        for i in range(iterations):
            self.log(f"\n--- Iteration {i+1}/{iterations} ---")
            
            # Test 1: Create then read
            result1 = self.test_create_then_read()
            all_results["create_then_read"].append(result1)
            self.results.append(result1)
            
            time.sleep(0.1)
            
            # Test 2: Add item then read
            result2 = self.test_write_then_read_items()
            all_results["write_then_read_items"].append(result2)
            self.results.append(result2)
            
            time.sleep(0.1)
            
            # Test 3: Rapid updates
            result3 = self.test_rapid_updates()
            all_results["rapid_updates"].append(result3)
            self.results.append(result3)
            
            time.sleep(0.2)
        
        # Calculate summary statistics
        self._print_summary(all_results)
        
        return all_results
    
    def _print_summary(self, all_results: Dict):
        """Print summary of consistency observations"""
        self.log("\n" + "=" * 70)
        self.log("CONSISTENCY ANALYSIS SUMMARY")
        self.log("=" * 70)
        
        # Test 1: Create then read
        create_read_results = all_results["create_then_read"]
        avg_consistency = sum(r["consistency_rate"] for r in create_read_results) / len(create_read_results)
        self.log(f"\nTest 1: Create Then Read")
        self.log(f"  Average consistency rate: {avg_consistency*100:.1f}%")
        self.log(f"  Eventual consistency delays observed: {avg_consistency < 1.0}")
        
        # Test 2: Add item then read
        write_read_results = all_results["write_then_read_items"]
        avg_item_consistency = sum(r["consistency_rate"] for r in write_read_results) / len(write_read_results)
        self.log(f"\nTest 2: Add Item Then Read")
        self.log(f"  Average item visibility rate: {avg_item_consistency*100:.1f}%")
        self.log(f"  Eventual consistency delays observed: {avg_item_consistency < 1.0}")
        
        # Test 3: Rapid updates
        rapid_results = all_results["rapid_updates"]
        avg_success = sum(r["updates_successful"] for r in rapid_results) / len(rapid_results)
        self.log(f"\nTest 3: Rapid Updates")
        self.log(f"  Average successful updates: {avg_success:.1f}/5")
        
        self.log("\n" + "=" * 70)
        self.log("KEY FINDINGS")
        self.log("=" * 70)
        
        if avg_consistency >= 0.99:
            self.log("✓ Read-after-write appears strongly consistent (>99%)")
        elif avg_consistency >= 0.9:
            self.log("⚠ Occasional eventual consistency delays observed")
        else:
            self.log("⚠ Significant eventual consistency delays observed")
        
        self.log(f"\nApplication impact:")
        if avg_consistency >= 0.95:
            self.log("  - Users unlikely to notice consistency delays")
        else:
            self.log("  - Users may occasionally see stale data")
            self.log("  - Recommendation: Implement client-side optimistic updates")
    
    def save_results(self, filename: str = "consistency_test_results.json"):
        """Save test results to JSON file"""
        try:
            with open(filename, 'w') as f:
                json.dump(self.results, f, indent=2)
            
            self.log(f"\n💾 Results saved to: {filename}")
            return True
            
        except Exception as e:
            self.log(f"❌ Error saving results: {e}")
            return False


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(
        description="DynamoDB Consistency Testing"
    )
    parser.add_argument(
        '--host',
        type=str,
        default='http://localhost:8080',
        help='Base URL of the API (default: http://localhost:8080)'
    )
    parser.add_argument(
        '--iterations',
        type=int,
        default=3,
        help='Number of test iterations (default: 3)'
    )
    parser.add_argument(
        '--output',
        type=str,
        default='consistency_test_results.json',
        help='Output filename (default: consistency_test_results.json)'
    )
    
    args = parser.parse_args()
    
    # Create tester instance
    tester = ConsistencyTester(args.host)
    
    # Run tests
    tester.run_all_tests(iterations=args.iterations)
    
    # Save results
    tester.save_results(args.output)


if __name__ == "__main__":
    main()

