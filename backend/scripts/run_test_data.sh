#!/bin/bash

set -e

# Change to the scripts directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../"

echo "Building test data generator..."
go build -o generate_test_data ./scripts/generate_test_data.go

echo "Running test data generator..."
./generate_test_data

echo "Test data generation completed!"