#!/bin/bash
# Check test coverage meets minimum threshold

set -e

THRESHOLD=80.0

echo "Running tests with coverage..."
go test -coverprofile=coverage.out ./... > /dev/null

echo "Generating coverage report..."
go tool cover -func=coverage.out > coverage.txt

# Extract total coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo "Total coverage: ${COVERAGE}%"
echo "Threshold: ${THRESHOLD}%"

# Compare coverage to threshold
if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
    echo "❌ Coverage ${COVERAGE}% is below threshold ${THRESHOLD}%"
    exit 1
else
    echo "✅ Coverage ${COVERAGE}% meets threshold ${THRESHOLD}%"
    exit 0
fi
