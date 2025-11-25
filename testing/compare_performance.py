"""
Compare PostgreSQL vs DynamoDB Performance Results

Analyzes mysql_test_results.json and dynamodb_test_results.json
to provide side-by-side comparison of performance metrics.

Usage:
    python compare_performance.py
"""

import json
from typing import List, Dict, Any


def load_results(filename: str) -> List[Dict[str, Any]]:
    """Load test results from JSON file"""
    try:
        with open(filename, 'r') as f:
            return json.load(f)
    except FileNotFoundError:
        print(f"❌ Error: {filename} not found")
        return []
    except json.JSONDecodeError:
        print(f"❌ Error: {filename} is not valid JSON")
        return []


def calculate_metrics(results: List[Dict[str, Any]], operation: str) -> Dict[str, float]:
    """Calculate performance metrics for a specific operation"""
    op_results = [r for r in results if r.get("operation") == operation and r.get("success", False)]
    
    if not op_results:
        return {
            "count": 0,
            "success_rate": 0,
            "avg": 0,
            "min": 0,
            "max": 0,
            "p50": 0,
            "p95": 0,
            "p99": 0
        }
    
    times = [r["response_time"] for r in op_results]
    sorted_times = sorted(times)
    
    return {
        "count": len(op_results),
        "success_rate": len(op_results) / 50 * 100,  # Out of 50 operations
        "avg": sum(times) / len(times),
        "min": min(times),
        "max": max(times),
        "p50": sorted_times[len(sorted_times) // 2],
        "p95": sorted_times[int(len(sorted_times) * 0.95)] if len(sorted_times) > 1 else sorted_times[0],
        "p99": sorted_times[int(len(sorted_times) * 0.99)] if len(sorted_times) > 1 else sorted_times[0]
    }


def print_comparison(pg_results: List[Dict], ddb_results: List[Dict]):
    """Print side-by-side comparison of results"""
    
    print("=" * 90)
    print(" " * 25 + "POSTGRESQL vs DYNAMODB COMPARISON")
    print("=" * 90)
    
    operations = [
        ("create_cart", "CREATE CART"),
        ("add_items", "ADD ITEMS"),
        ("get_cart", "GET CART")
    ]
    
    for op_key, op_name in operations:
        pg_metrics = calculate_metrics(pg_results, op_key)
        ddb_metrics = calculate_metrics(ddb_results, op_key)
        
        print(f"\n{op_name}:")
        print("-" * 90)
        print(f"{'Metric':<20} {'PostgreSQL':<25} {'DynamoDB':<25} {'Winner':<20}")
        print("-" * 90)
        
        # Success Rate
        print(f"{'Success Rate':<20} {pg_metrics['success_rate']:<24.1f}% {ddb_metrics['success_rate']:<24.1f}% ", end="")
        if pg_metrics['success_rate'] > ddb_metrics['success_rate']:
            print("PostgreSQL ✓")
        elif ddb_metrics['success_rate'] > pg_metrics['success_rate']:
            print("DynamoDB ✓")
        else:
            print("Tie")
        
        # Average Response Time
        print(f"{'Avg Response':<20} {pg_metrics['avg']:<24.2f}ms {ddb_metrics['avg']:<24.2f}ms ", end="")
        if pg_metrics['avg'] < ddb_metrics['avg']:
            print(f"PostgreSQL ✓ ({ddb_metrics['avg'] - pg_metrics['avg']:.2f}ms faster)")
        elif ddb_metrics['avg'] < pg_metrics['avg']:
            print(f"DynamoDB ✓ ({pg_metrics['avg'] - ddb_metrics['avg']:.2f}ms faster)")
        else:
            print("Tie")
        
        # Min Response Time
        print(f"{'Min Response':<20} {pg_metrics['min']:<24.2f}ms {ddb_metrics['min']:<24.2f}ms ", end="")
        if pg_metrics['min'] < ddb_metrics['min']:
            print("PostgreSQL ✓")
        elif ddb_metrics['min'] < pg_metrics['min']:
            print("DynamoDB ✓")
        else:
            print("Tie")
        
        # Max Response Time
        print(f"{'Max Response':<20} {pg_metrics['max']:<24.2f}ms {ddb_metrics['max']:<24.2f}ms ", end="")
        if pg_metrics['max'] < ddb_metrics['max']:
            print("PostgreSQL ✓")
        elif ddb_metrics['max'] < pg_metrics['max']:
            print("DynamoDB ✓")
        else:
            print("Tie")
        
        # P50 (Median)
        print(f"{'P50 (Median)':<20} {pg_metrics['p50']:<24.2f}ms {ddb_metrics['p50']:<24.2f}ms ", end="")
        if pg_metrics['p50'] < ddb_metrics['p50']:
            print("PostgreSQL ✓")
        elif ddb_metrics['p50'] < pg_metrics['p50']:
            print("DynamoDB ✓")
        else:
            print("Tie")
        
        # P95
        print(f"{'P95':<20} {pg_metrics['p95']:<24.2f}ms {ddb_metrics['p95']:<24.2f}ms ", end="")
        if pg_metrics['p95'] < ddb_metrics['p95']:
            print("PostgreSQL ✓")
        elif ddb_metrics['p95'] < pg_metrics['p95']:
            print("DynamoDB ✓")
        else:
            print("Tie")
        
        # P99
        print(f"{'P99':<20} {pg_metrics['p99']:<24.2f}ms {ddb_metrics['p99']:<24.2f}ms ", end="")
        if pg_metrics['p99'] < ddb_metrics['p99']:
            print("PostgreSQL ✓")
        elif ddb_metrics['p99'] < pg_metrics['p99']:
            print("DynamoDB ✓")
        else:
            print("Tie")
    
    # Overall summary
    print("\n" + "=" * 90)
    print("OVERALL SUMMARY")
    print("=" * 90)
    
    # Calculate total successful operations
    pg_total_success = sum(1 for r in pg_results if r.get("success", False))
    ddb_total_success = sum(1 for r in ddb_results if r.get("success", False))
    
    print(f"\nTotal Successful Operations:")
    print(f"  PostgreSQL: {pg_total_success}/150 ({pg_total_success/150*100:.1f}%)")
    print(f"  DynamoDB:   {ddb_total_success}/150 ({ddb_total_success/150*100:.1f}%)")
    
    # Calculate average response time across all operations
    pg_all_times = [r["response_time"] for r in pg_results if r.get("success", False)]
    ddb_all_times = [r["response_time"] for r in ddb_results if r.get("success", False)]
    
    if pg_all_times and ddb_all_times:
        pg_overall_avg = sum(pg_all_times) / len(pg_all_times)
        ddb_overall_avg = sum(ddb_all_times) / len(ddb_all_times)
        
        print(f"\nOverall Average Response Time:")
        print(f"  PostgreSQL: {pg_overall_avg:.2f}ms")
        print(f"  DynamoDB:   {ddb_overall_avg:.2f}ms")
        
        if pg_overall_avg < ddb_overall_avg:
            diff = ddb_overall_avg - pg_overall_avg
            print(f"\n✅ Winner: PostgreSQL (faster by {diff:.2f}ms on average)")
        elif ddb_overall_avg < pg_overall_avg:
            diff = pg_overall_avg - ddb_overall_avg
            print(f"\n✅ Winner: DynamoDB (faster by {diff:.2f}ms on average)")
        else:
            print(f"\n🤝 Performance is equivalent")
    
    print("\n" + "=" * 90)


def main():
    """Main entry point"""
    print("\n🔍 Loading test results...")
    
    pg_results = load_results("mysql_test_results.json")
    ddb_results = load_results("dynamodb_test_results.json")
    
    if not pg_results:
        print("⚠️  PostgreSQL results not found or empty")
        print("   Run: python cart_performance_test.py --host <your-ip>:8080")
    
    if not ddb_results:
        print("⚠️  DynamoDB results not found or empty")
        print("   Run: python dynamodb_performance_test.py --host <your-ip>:8080")
    
    if not pg_results or not ddb_results:
        print("\n❌ Cannot proceed without both result files")
        return
    
    print(f"✅ Loaded {len(pg_results)} PostgreSQL results")
    print(f"✅ Loaded {len(ddb_results)} DynamoDB results\n")
    
    print_comparison(pg_results, ddb_results)


if __name__ == "__main__":
    main()

