#!/bin/bash
set -e

echo "Running unit tests..."
go test ./test/unit/...

echo "Running integration tests..."
./scripts/integration_test.sh

echo "All tests passed!"