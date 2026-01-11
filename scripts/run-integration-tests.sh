#!/bin/bash

# Integration Test Runner for Connect 4 Multiplayer System
# This script runs comprehensive integration tests with cloud services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if .env file exists
if [ ! -f ".env" ]; then
    print_error ".env file not found. Please create it with your cloud service credentials."
    print_status "See .env.example for required variables."
    exit 1
fi

# Load environment variables
print_status "Loading environment variables..."
source .env

# Check required environment variables
check_env_var() {
    if [ -z "${!1}" ]; then
        print_error "Environment variable $1 is not set"
        return 1
    fi
    return 0
}

print_status "Checking required environment variables..."

# Check database configuration
if ! check_env_var "DATABASE_URL"; then
    print_error "DATABASE_URL is required for integration tests"
    exit 1
fi

# Check Kafka configuration (optional for some tests)
if [ -z "$KAFKA_BOOTSTRAP_SERVERS" ] || [ -z "$KAFKA_API_KEY" ] || [ -z "$KAFKA_API_SECRET" ]; then
    print_warning "Kafka configuration incomplete. Some tests may be skipped."
fi

print_success "Environment variables validated"

# Set test-specific environment variables
export ENVIRONMENT=test
export TEST_DATABASE_URL=${TEST_DATABASE_URL:-$DATABASE_URL}
export KAFKA_TOPIC=test-game-events
export KAFKA_CONSUMER_GROUP=test-analytics-service
export SERVER_PORT=0

print_status "Test environment configured"

# Build the application
print_status "Building application..."
if ! go build -o bin/test-server cmd/server/main.go; then
    print_error "Failed to build server"
    exit 1
fi

if ! go build -o bin/test-analytics cmd/analytics/main.go; then
    print_error "Failed to build analytics service"
    exit 1
fi

print_success "Application built successfully"

# Run database migrations
print_status "Running database migrations..."
if ! go run cmd/migrate/main.go; then
    print_error "Database migrations failed"
    exit 1
fi

print_success "Database migrations completed"

# Function to run specific test suite
run_test_suite() {
    local test_name=$1
    local test_pattern=$2
    local timeout=${3:-10m}
    
    print_status "Running $test_name..."
    
    if go test -tags=integration -timeout=$timeout -v ./tests/integration -run="$test_pattern" 2>&1 | tee "test-results-$test_name.log"; then
        print_success "$test_name completed successfully"
        return 0
    else
        print_error "$test_name failed"
        return 1
    fi
}

# Parse command line arguments
TEST_SUITE=${1:-"all"}
VERBOSE=${2:-"false"}

case $TEST_SUITE in
    "e2e"|"end-to-end")
        print_status "Running End-to-End Integration Tests"
        run_test_suite "e2e" "TestE2ETestSuite" "15m"
        ;;
    "performance"|"perf")
        print_status "Running Performance Tests"
        run_test_suite "performance" "TestPerformanceTestSuite" "20m"
        ;;
    "kafka")
        print_status "Running Kafka Integration Tests"
        if [ -z "$KAFKA_API_KEY" ]; then
            print_error "Kafka credentials required for Kafka tests"
            exit 1
        fi
        run_test_suite "kafka" "TestKafkaIntegration" "10m"
        ;;
    "all"|*)
        print_status "Running All Integration Tests"
        
        # Run E2E tests
        if ! run_test_suite "e2e" "TestE2ETestSuite" "15m"; then
            print_error "E2E tests failed"
            exit 1
        fi
        
        # Run Performance tests
        if ! run_test_suite "performance" "TestPerformanceTestSuite" "20m"; then
            print_error "Performance tests failed"
            exit 1
        fi
        
        # Run Kafka tests if credentials available
        if [ -n "$KAFKA_API_KEY" ] && [ -n "$KAFKA_API_SECRET" ]; then
            if ! run_test_suite "kafka" "TestKafkaIntegration" "10m"; then
                print_warning "Kafka tests failed, but continuing..."
            fi
        else
            print_warning "Skipping Kafka tests - credentials not available"
        fi
        ;;
esac

# Generate test report
print_status "Generating test report..."

cat > integration-test-report.md << EOF
# Integration Test Report

**Date:** $(date)
**Environment:** $ENVIRONMENT
**Database:** $(echo $DATABASE_URL | sed 's/:[^@]*@/:***@/')
**Kafka:** $(echo $KAFKA_BOOTSTRAP_SERVERS | cut -d: -f1):***

## Test Results

EOF

# Add results from log files
for log_file in test-results-*.log; do
    if [ -f "$log_file" ]; then
        test_name=$(basename "$log_file" .log | sed 's/test-results-//')
        echo "### $test_name" >> integration-test-report.md
        echo '```' >> integration-test-report.md
        tail -20 "$log_file" >> integration-test-report.md
        echo '```' >> integration-test-report.md
        echo "" >> integration-test-report.md
    fi
done

cat >> integration-test-report.md << EOF

## Performance Metrics

- **Database Connection Pool:** ${DATABASE_MAX_OPEN_CONNS:-25} max connections
- **Test Duration:** $(date)
- **Cloud Services:** Supabase PostgreSQL, Confluent Cloud Kafka

## Recommendations

1. Monitor database connection usage in production
2. Set up alerts for WebSocket connection limits
3. Implement graceful degradation for high load scenarios
4. Consider horizontal scaling for analytics processing

EOF

print_success "Integration test report generated: integration-test-report.md"

# Clean up test artifacts
print_status "Cleaning up test artifacts..."
rm -f bin/test-server bin/test-analytics
rm -f test-results-*.log

print_success "Integration tests completed successfully!"

# Display summary
echo ""
echo "=========================================="
echo "         INTEGRATION TEST SUMMARY"
echo "=========================================="
echo "Test Suite: $TEST_SUITE"
echo "Environment: $ENVIRONMENT"
echo "Database: Connected"
echo "Kafka: $([ -n "$KAFKA_API_KEY" ] && echo "Connected" || echo "Skipped")"
echo "Report: integration-test-report.md"
echo "=========================================="