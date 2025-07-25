#!/usr/bin/env python3
"""
Entity State Events Integration Tests Runner
Runs only the entity state events-related integration tests.
"""

import sys
import subprocess
import os

def main():
    """Run entity state events integration tests."""
    print("Starting entity state events integration tests...")
    
    # Set environment variables
    os.environ.setdefault("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
    os.environ.setdefault("CI", os.environ.get("CI", ""))
    
    # Run entity state events tests
    test_files = [
        "test_entity_state_events_collection.py",
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
        print("\n🎉 All entity state events tests passed!")
    else:
        print(f"\n💥 Entity state events tests failed with exit code {exit_code}")
    
    sys.exit(exit_code)

if __name__ == "__main__":
    main()
