#!/usr/bin/env python3
"""
Main Integration Tests Runner
Supports running specific test suites or all tests based on TEST_SUITE environment variable.
"""

import sys
import os
import subprocess

def main():
    """Run integration tests based on TEST_SUITE environment variable."""
    test_suite = os.environ.get("TEST_SUITE", "all").lower()
    
    print(f"🚀 Starting integration tests for suite: {test_suite}")
    
    # Map test suites to their runner scripts
    test_runners = {
        "metrics": "./run_metrics_tests.py",
        "logs": "./run_logs_tests.py", 
        "events": "./run_events_tests.py",
        "manifests": "./run_manifests_tests.py",
        "entity-state": "./run_entity_state_tests.py",
        "entity_state": "./run_entity_state_tests.py",  # Alternative naming
        "all": None  # Run all tests with pytest
    }
    
    if test_suite not in test_runners:
        print(f"❌ Unknown test suite: {test_suite}")
        print(f"Available test suites: {', '.join(test_runners.keys())}")
        sys.exit(1)
    
    # Set common environment variables
    os.environ.setdefault("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
    os.environ.setdefault("CI", os.environ.get("CI", ""))
    
    if test_suite == "all":
        # Run all tests with pytest (default behavior)
        print("Running full test suite...")
        result = subprocess.run([
            "python", "-m", "pytest", 
            "-s", 
            "--tb=short"
        ], cwd="/app")
        
        if result.returncode == 0:
            print("\n🎉 All integration tests passed!")
        else:
            print(f"\n💥 Integration tests failed with exit code {result.returncode}")
        
        sys.exit(result.returncode)
    else:
        # Run specific test suite
        runner_script = test_runners[test_suite]
        print(f"Running {test_suite} test suite using {runner_script}...")
        
        result = subprocess.run(["python", runner_script], cwd="/app")
        sys.exit(result.returncode)

if __name__ == "__main__":
    main()
