#!/bin/bash
#
# Quick-start script to generate entity state events and relationships JSON files
#
# This script:
# 1. Checks if clickhouse-connect is installed
# 2. Checks if port-forward is active (via Skaffold)
# 3. Runs the Python script to generate JSON files
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== Generate Entity State Events & Relationships Script ==="
echo

# Check if Python is available
if ! command -v python3 &> /dev/null; then
    echo "Error: python3 is not installed"
    exit 1
fi

# Check if clickhouse-connect is installed
echo "Checking for clickhouse-connect..."
if ! python3 -c "import clickhouse_connect" 2>/dev/null; then
    echo "clickhouse-connect is not installed."
    echo
    echo "Install it with:"
    echo "  pip3 install --user -r requirements-relationships.txt"
    echo
    echo "Or:"
    echo "  pip3 install --user clickhouse-connect"
    echo
    exit 1
fi

echo "✓ clickhouse-connect is installed"
echo

# Check if ClickHouse is accessible (via Skaffold port-forward)
echo "Checking ClickHouse connection (expecting Skaffold port-forward)..."
if ! python3 -c "import clickhouse_connect; clickhouse_connect.get_client(host='localhost', port=8123, connect_timeout=2)" 2>/dev/null; then
    echo "Cannot connect to ClickHouse at localhost:8123"
    echo
    echo "Make sure Skaffold is running with port-forwarding enabled:"
    echo "  skaffold dev"
    echo
    echo "Skaffold should automatically forward ClickHouse HTTP port 8123."
    exit 1
fi

echo "✓ Connected to ClickHouse"
echo

# Run the script
echo "Generating entity state events and relationships..."
python3 generate_relationships.py "$@"

echo
echo "✓ Done!"
