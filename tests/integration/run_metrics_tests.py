#!/usr/bin/env python3
"""
Metrics Integration Tests Runner
Runs only the metrics-related integration tests.
"""

import sys
import subprocess
import os

def main():
    """Run metrics integration tests."""
    print("Starting metrics integration tests...")
    
    # Set environment variables
    os.environ.setdefault("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
    os.environ.setdefault("CI", os.environ.get("CI", ""))
    
    # Run metrics tests
    test_files = [
        "test_metric_collection.py",
    ]
    
    exit_code = 0
    for test_file in test_files:
        print(f"\n=== Running {test_file} ===")
        result = subprocess.run([
            "python", "-m", "pytest", 
            test_file, 
            "-s", 
            "--tb=short"
        ], cwd="/app")
        
        if result.returncode != 0:
            print(f"❌ {test_file} failed with exit code {result.returncode}")
            exit_code = result.returncode
        else:
            print(f"✅ {test_file} passed")
    
    if exit_code == 0:
        print("\n🎉 All metrics tests passed!")
    else:
        print(f"\n💥 Metrics tests failed with exit code {exit_code}")
    
    sys.exit(exit_code)

if __name__ == "__main__":
    main()
