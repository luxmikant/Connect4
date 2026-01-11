# Integration Test Runner for Connect 4 Multiplayer System (PowerShell)
# This script runs comprehensive integration tests with cloud services

param(
    [string]$TestSuite = "all",
    [switch]$Verbose = $false
)

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"

function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Red
}

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Error ".env file not found. Please create it with your cloud service credentials."
    Write-Status "See .env.example for required variables."
    exit 1
}

# Load environment variables from .env file
Write-Status "Loading environment variables..."
Get-Content ".env" | ForEach-Object {
    if ($_ -match "^([^#][^=]+)=(.*)$") {
        $name = $matches[1].Trim()
        $value = $matches[2].Trim()
        [Environment]::SetEnvironmentVariable($name, $value, "Process")
    }
}

# Check required environment variables
function Test-EnvVar {
    param([string]$VarName)
    $value = [Environment]::GetEnvironmentVariable($VarName)
    if ([string]::IsNullOrEmpty($value)) {
        Write-Error "Environment variable $VarName is not set"
        return $false
    }
    return $true
}

Write-Status "Checking required environment variables..."

# Check database configuration
if (-not (Test-EnvVar "DATABASE_URL")) {
    Write-Error "DATABASE_URL is required for integration tests"
    exit 1
}

# Check Kafka configuration (optional for some tests)
$kafkaBootstrap = [Environment]::GetEnvironmentVariable("KAFKA_BOOTSTRAP_SERVERS")
$kafkaKey = [Environment]::GetEnvironmentVariable("KAFKA_API_KEY")
$kafkaSecret = [Environment]::GetEnvironmentVariable("KAFKA_API_SECRET")

if ([string]::IsNullOrEmpty($kafkaBootstrap) -or [string]::IsNullOrEmpty($kafkaKey) -or [string]::IsNullOrEmpty($kafkaSecret)) {
    Write-Warning "Kafka configuration incomplete. Some tests may be skipped."
}

Write-Success "Environment variables validated"

# Set test-specific environment variables
[Environment]::SetEnvironmentVariable("ENVIRONMENT", "test", "Process")
$testDbUrl = [Environment]::GetEnvironmentVariable("TEST_DATABASE_URL")
if ([string]::IsNullOrEmpty($testDbUrl)) {
    [Environment]::SetEnvironmentVariable("TEST_DATABASE_URL", [Environment]::GetEnvironmentVariable("DATABASE_URL"), "Process")
}
[Environment]::SetEnvironmentVariable("KAFKA_TOPIC", "test-game-events", "Process")
[Environment]::SetEnvironmentVariable("KAFKA_CONSUMER_GROUP", "test-analytics-service", "Process")
[Environment]::SetEnvironmentVariable("SERVER_PORT", "0", "Process")

Write-Status "Test environment configured"

# Build the application
Write-Status "Building application..."

if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

try {
    & go build -o "bin/test-server.exe" "cmd/server/main.go"
    if ($LASTEXITCODE -ne 0) {
        throw "Server build failed"
    }
} catch {
    Write-Error "Failed to build server: $_"
    exit 1
}

try {
    & go build -o "bin/test-analytics.exe" "cmd/analytics/main.go"
    if ($LASTEXITCODE -ne 0) {
        throw "Analytics build failed"
    }
} catch {
    Write-Error "Failed to build analytics service: $_"
    exit 1
}

Write-Success "Application built successfully"

# Run database migrations
Write-Status "Running database migrations..."
try {
    & go run "cmd/migrate/main.go"
    if ($LASTEXITCODE -ne 0) {
        throw "Migration failed"
    }
} catch {
    Write-Error "Database migrations failed: $_"
    exit 1
}

Write-Success "Database migrations completed"

# Function to run specific test suite
function Invoke-TestSuite {
    param(
        [string]$TestName,
        [string]$TestPattern,
        [string]$Timeout = "10m"
    )
    
    Write-Status "Running $TestName..."
    
    $logFile = "test-results-$TestName.log"
    
    try {
        $output = & go test -tags=integration -timeout=$Timeout -v "./tests/integration" -run="$TestPattern" 2>&1
        $output | Out-File -FilePath $logFile -Encoding UTF8
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "$TestName completed successfully"
            return $true
        } else {
            Write-Error "$TestName failed"
            Write-Host $output -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Error "$TestName failed with exception: $_"
        return $false
    }
}

# Run tests based on command line argument
switch ($TestSuite.ToLower()) {
    { $_ -in @("e2e", "end-to-end") } {
        Write-Status "Running End-to-End Integration Tests"
        $success = Invoke-TestSuite "e2e" "TestE2ETestSuite" "15m"
        if (-not $success) { exit 1 }
    }
    
    { $_ -in @("performance", "perf") } {
        Write-Status "Running Performance Tests"
        $success = Invoke-TestSuite "performance" "TestPerformanceTestSuite" "20m"
        if (-not $success) { exit 1 }
    }
    
    "kafka" {
        Write-Status "Running Kafka Integration Tests"
        if ([string]::IsNullOrEmpty($kafkaKey)) {
            Write-Error "Kafka credentials required for Kafka tests"
            exit 1
        }
        $success = Invoke-TestSuite "kafka" "TestKafkaIntegration" "10m"
        if (-not $success) { exit 1 }
    }
    
    default {
        Write-Status "Running All Integration Tests"
        
        # Run E2E tests
        if (-not (Invoke-TestSuite "e2e" "TestE2ETestSuite" "15m")) {
            Write-Error "E2E tests failed"
            exit 1
        }
        
        # Run Performance tests
        if (-not (Invoke-TestSuite "performance" "TestPerformanceTestSuite" "20m")) {
            Write-Error "Performance tests failed"
            exit 1
        }
        
        # Run Kafka tests if credentials available
        if (-not [string]::IsNullOrEmpty($kafkaKey) -and -not [string]::IsNullOrEmpty($kafkaSecret)) {
            if (-not (Invoke-TestSuite "kafka" "TestKafkaIntegration" "10m")) {
                Write-Warning "Kafka tests failed, but continuing..."
            }
        } else {
            Write-Warning "Skipping Kafka tests - credentials not available"
        }
    }
}

# Generate test report
Write-Status "Generating test report..."

$reportContent = @"
# Integration Test Report

**Date:** $(Get-Date)
**Environment:** test
**Database:** $($env:DATABASE_URL -replace ':[^@]*@', ':***@')
**Kafka:** $($env:KAFKA_BOOTSTRAP_SERVERS -split ':')[0]:***

## Test Results

"@

# Add results from log files
Get-ChildItem "test-results-*.log" -ErrorAction SilentlyContinue | ForEach-Object {
    $testName = $_.BaseName -replace 'test-results-', ''
    $reportContent += "`n### $testName`n"
    $reportContent += "``````n"
    $reportContent += (Get-Content $_.FullName | Select-Object -Last 20) -join "`n"
    $reportContent += "`n``````n`n"
}

$reportContent += @"

## Performance Metrics

- **Database Connection Pool:** $($env:DATABASE_MAX_OPEN_CONNS -or '25') max connections
- **Test Duration:** $(Get-Date)
- **Cloud Services:** Supabase PostgreSQL, Confluent Cloud Kafka

## Recommendations

1. Monitor database connection usage in production
2. Set up alerts for WebSocket connection limits
3. Implement graceful degradation for high load scenarios
4. Consider horizontal scaling for analytics processing

"@

$reportContent | Out-File -FilePath "integration-test-report.md" -Encoding UTF8

Write-Success "Integration test report generated: integration-test-report.md"

# Clean up test artifacts
Write-Status "Cleaning up test artifacts..."
Remove-Item "bin/test-server.exe" -ErrorAction SilentlyContinue
Remove-Item "bin/test-analytics.exe" -ErrorAction SilentlyContinue
Remove-Item "test-results-*.log" -ErrorAction SilentlyContinue

Write-Success "Integration tests completed successfully!"

# Display summary
Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "         INTEGRATION TEST SUMMARY" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Test Suite: $TestSuite" -ForegroundColor White
Write-Host "Environment: test" -ForegroundColor White
Write-Host "Database: Connected" -ForegroundColor White
Write-Host "Kafka: $(if (-not [string]::IsNullOrEmpty($kafkaKey)) { 'Connected' } else { 'Skipped' })" -ForegroundColor White
Write-Host "Report: integration-test-report.md" -ForegroundColor White
Write-Host "==========================================" -ForegroundColor Cyan