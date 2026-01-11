#!/bin/bash

# Connect 4 Multiplayer - Credential Setup Script
# This script helps you set up and validate your cloud service credentials

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to validate environment file
validate_env_file() {
    local env_file="$1"
    
    if [[ ! -f "$env_file" ]]; then
        print_error "Environment file $env_file not found"
        return 1
    fi
    
    print_status "Validating environment file: $env_file"
    
    # Source the env file
    set -a
    source "$env_file"
    set +a
    
    # Check required variables
    local required_vars=(
        "ENVIRONMENT"
        "DATABASE_URL"
        "KAFKA_BOOTSTRAP_SERVERS"
    )
    
    local missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var}" ]]; then
            missing_vars+=("$var")
        fi
    done
    
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        print_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        return 1
    fi
    
    print_success "Environment file validation passed"
    return 0
}

# Function to test database connection
test_database_connection() {
    print_status "Testing database connection..."
    
    if [[ -z "$DATABASE_URL" ]]; then
        print_error "DATABASE_URL not set"
        return 1
    fi
    
    # Check if psql is available
    if ! command_exists psql; then
        print_warning "psql not found. Install PostgreSQL client to test database connection"
        print_status "You can install it with:"
        echo "  - Ubuntu/Debian: sudo apt-get install postgresql-client"
        echo "  - macOS: brew install postgresql"
        echo "  - Windows: Download from https://www.postgresql.org/download/"
        return 0
    fi
    
    # Test connection with a simple query
    if psql "$DATABASE_URL" -c "SELECT version();" >/dev/null 2>&1; then
        print_success "Database connection successful"
        
        # Get database info
        local db_version=$(psql "$DATABASE_URL" -t -c "SELECT version();" 2>/dev/null | head -1 | xargs)
        print_status "Database version: $db_version"
        
        return 0
    else
        print_error "Database connection failed"
        print_status "Please check your DATABASE_URL and ensure:"
        echo "  1. The database server is running"
        echo "  2. The credentials are correct"
        echo "  3. The database exists"
        echo "  4. Network connectivity is available"
        return 1
    fi
}

# Function to test Kafka connection
test_kafka_connection() {
    print_status "Testing Kafka connection..."
    
    if [[ -z "$KAFKA_BOOTSTRAP_SERVERS" ]]; then
        print_error "KAFKA_BOOTSTRAP_SERVERS not set"
        return 1
    fi
    
    # For Confluent Cloud, we need API key and secret
    if [[ "$KAFKA_BOOTSTRAP_SERVERS" == *"confluent.cloud"* ]]; then
        if [[ -z "$KAFKA_API_KEY" || -z "$KAFKA_API_SECRET" ]]; then
            print_error "Confluent Cloud requires KAFKA_API_KEY and KAFKA_API_SECRET"
            return 1
        fi
        
        print_status "Confluent Cloud configuration detected"
        print_status "Bootstrap servers: $KAFKA_BOOTSTRAP_SERVERS"
        print_status "API Key: ${KAFKA_API_KEY:0:8}..."
        
        # Check if confluent CLI is available
        if command_exists confluent; then
            print_status "Testing with Confluent CLI..."
            # Note: This would require additional setup with confluent CLI
            print_warning "Manual verification recommended via Confluent Cloud dashboard"
        else
            print_warning "Confluent CLI not found. Install it for advanced testing:"
            echo "  curl -sL --http1.1 https://cnfl.io/cli | sh -s -- latest"
        fi
    else
        print_status "Local Kafka configuration detected"
        print_status "Bootstrap servers: $KAFKA_BOOTSTRAP_SERVERS"
        
        # For local Kafka, we can try a simple connection test
        if command_exists kafka-topics; then
            if kafka-topics --bootstrap-server "$KAFKA_BOOTSTRAP_SERVERS" --list >/dev/null 2>&1; then
                print_success "Kafka connection successful"
            else
                print_error "Kafka connection failed"
                return 1
            fi
        else
            print_warning "kafka-topics command not found. Cannot test Kafka connection"
            print_status "Ensure Kafka is running: docker-compose up kafka"
        fi
    fi
    
    return 0
}

# Function to run database migrations
run_migrations() {
    print_status "Running database migrations..."
    
    if [[ ! -f "cmd/migrate/main.go" ]]; then
        print_error "Migration tool not found at cmd/migrate/main.go"
        return 1
    fi
    
    if go run cmd/migrate/main.go; then
        print_success "Database migrations completed successfully"
    else
        print_error "Database migrations failed"
        return 1
    fi
}

# Function to create .env file from template
create_env_file() {
    if [[ -f ".env" ]]; then
        print_warning ".env file already exists"
        read -p "Do you want to overwrite it? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status "Keeping existing .env file"
            return 0
        fi
    fi
    
    if [[ -f ".env.example" ]]; then
        cp .env.example .env
        print_success "Created .env file from .env.example"
        print_status "Please edit .env file with your actual credentials"
        print_status "See docs/cloud-setup-guide.md for detailed instructions"
    else
        print_error ".env.example file not found"
        return 1
    fi
}

# Function to show setup instructions
show_setup_instructions() {
    echo
    print_status "=== Connect 4 Multiplayer Credential Setup ==="
    echo
    print_status "This script helps you set up and validate cloud service credentials."
    echo
    print_status "Prerequisites:"
    echo "  1. Supabase account and project"
    echo "  2. Confluent Cloud account and cluster (optional)"
    echo "  3. .env file with your credentials"
    echo
    print_status "For detailed setup instructions, see:"
    echo "  docs/cloud-setup-guide.md"
    echo
}

# Function to show menu
show_menu() {
    echo
    print_status "What would you like to do?"
    echo "  1) Create .env file from template"
    echo "  2) Validate environment configuration"
    echo "  3) Test database connection"
    echo "  4) Test Kafka connection"
    echo "  5) Run database migrations"
    echo "  6) Run all tests"
    echo "  7) Show setup instructions"
    echo "  8) Exit"
    echo
}

# Main function
main() {
    show_setup_instructions
    
    while true; do
        show_menu
        read -p "Enter your choice (1-8): " choice
        
        case $choice in
            1)
                create_env_file
                ;;
            2)
                if validate_env_file ".env"; then
                    print_success "Environment configuration is valid"
                else
                    print_error "Environment configuration validation failed"
                fi
                ;;
            3)
                if validate_env_file ".env"; then
                    test_database_connection
                fi
                ;;
            4)
                if validate_env_file ".env"; then
                    test_kafka_connection
                fi
                ;;
            5)
                if validate_env_file ".env"; then
                    run_migrations
                fi
                ;;
            6)
                if validate_env_file ".env"; then
                    test_database_connection
                    test_kafka_connection
                    run_migrations
                    print_success "All tests completed"
                fi
                ;;
            7)
                show_setup_instructions
                ;;
            8)
                print_status "Goodbye!"
                exit 0
                ;;
            *)
                print_error "Invalid choice. Please enter 1-8."
                ;;
        esac
        
        echo
        read -p "Press Enter to continue..."
    done
}

# Run main function
main "$@"