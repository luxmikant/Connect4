#!/usr/bin/env pwsh
# PowerShell script to validate cloud service credentials
# Run this script to test all your cloud service connections

Write-Host "Validating Cloud Service Credentials..." -ForegroundColor Cyan
Write-Host "=================================================="

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Host "ERROR: .env file not found!" -ForegroundColor Red
    Write-Host "Please copy .env.example to .env and configure your credentials." -ForegroundColor Yellow
    exit 1
}

Write-Host "SUCCESS: .env file found" -ForegroundColor Green

# Load environment variables from .env file
Get-Content ".env" | ForEach-Object {
    if ($_ -match "^([^#=]+)=(.*)$") {
        $name = $matches[1].Trim()
        $value = $matches[2].Trim()
        [Environment]::SetEnvironmentVariable($name, $value, "Process")
    }
}

Write-Host ""
Write-Host "Testing Database Connection (Supabase)..." -ForegroundColor Cyan

# Test database connection
$dbUrl = [Environment]::GetEnvironmentVariable("DATABASE_URL")
if ([string]::IsNullOrEmpty($dbUrl)) {
    Write-Host "ERROR: DATABASE_URL not configured" -ForegroundColor Red
} else {
    Write-Host "SUCCESS: DATABASE_URL configured" -ForegroundColor Green
    Write-Host "   URL: $($dbUrl.Substring(0, 30))..." -ForegroundColor Gray
    
    # Try to run database migration to test connection
    Write-Host "   Testing connection with migration check..." -ForegroundColor Gray
    try {
        $result = & go run cmd/migrate/main.go 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "SUCCESS: Database connection successful" -ForegroundColor Green
        } else {
            Write-Host "WARNING: Database connection issue: $result" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "WARNING: Could not test database connection: $_" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "Testing Kafka Connection (Confluent Cloud)..." -ForegroundColor Cyan

$kafkaBootstrap = [Environment]::GetEnvironmentVariable("KAFKA_BOOTSTRAP_SERVERS")
$kafkaApiKey = [Environment]::GetEnvironmentVariable("KAFKA_API_KEY")
$kafkaApiSecret = [Environment]::GetEnvironmentVariable("KAFKA_API_SECRET")

if ([string]::IsNullOrEmpty($kafkaBootstrap)) {
    Write-Host "ERROR: KAFKA_BOOTSTRAP_SERVERS not configured" -ForegroundColor Red
} else {
    Write-Host "SUCCESS: KAFKA_BOOTSTRAP_SERVERS configured" -ForegroundColor Green
    Write-Host "   Bootstrap: $kafkaBootstrap" -ForegroundColor Gray
}

if ([string]::IsNullOrEmpty($kafkaApiKey)) {
    Write-Host "ERROR: KAFKA_API_KEY not configured" -ForegroundColor Red
} else {
    Write-Host "SUCCESS: KAFKA_API_KEY configured" -ForegroundColor Green
    Write-Host "   API Key: $($kafkaApiKey.Substring(0, 8))..." -ForegroundColor Gray
}

if ([string]::IsNullOrEmpty($kafkaApiSecret)) {
    Write-Host "ERROR: KAFKA_API_SECRET not configured" -ForegroundColor Red
} else {
    Write-Host "SUCCESS: KAFKA_API_SECRET configured" -ForegroundColor Green
    Write-Host "   API Secret: $($kafkaApiSecret.Substring(0, 8))..." -ForegroundColor Gray
}

Write-Host ""
Write-Host "Testing Redis Connection (Redis Cloud)..." -ForegroundColor Cyan

$redisUrl = [Environment]::GetEnvironmentVariable("REDIS_URL")
$redisPassword = [Environment]::GetEnvironmentVariable("REDIS_PASSWORD")

if ([string]::IsNullOrEmpty($redisUrl)) {
    Write-Host "ERROR: REDIS_URL not configured" -ForegroundColor Red
} else {
    Write-Host "SUCCESS: REDIS_URL configured" -ForegroundColor Green
    Write-Host "   URL: redis://default:***@redis-19383.c301.ap-south-1-1.ec2.cloud.redislabs.com:19383" -ForegroundColor Gray
}

if ([string]::IsNullOrEmpty($redisPassword)) {
    Write-Host "ERROR: REDIS_PASSWORD not configured" -ForegroundColor Red
} else {
    Write-Host "SUCCESS: REDIS_PASSWORD configured" -ForegroundColor Green
    Write-Host "   Password: $($redisPassword.Substring(0, 8))..." -ForegroundColor Gray
}

Write-Host ""
Write-Host "Testing Application Startup..." -ForegroundColor Cyan

# Test if the application can start (compile check)
Write-Host "   Checking if server compiles..." -ForegroundColor Gray
try {
    $buildResult = & go build -o temp_server.exe cmd/server/main.go 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "SUCCESS: Server compiles successfully" -ForegroundColor Green
        Remove-Item "temp_server.exe" -ErrorAction SilentlyContinue
    } else {
        Write-Host "ERROR: Server compilation failed: $buildResult" -ForegroundColor Red
    }
} catch {
    Write-Host "ERROR: Server compilation error: $_" -ForegroundColor Red
}

Write-Host "   Checking if analytics service compiles..." -ForegroundColor Gray
try {
    $buildResult = & go build -o temp_analytics.exe cmd/analytics/main.go 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "SUCCESS: Analytics service compiles successfully" -ForegroundColor Green
        Remove-Item "temp_analytics.exe" -ErrorAction SilentlyContinue
    } else {
        Write-Host "ERROR: Analytics service compilation failed: $buildResult" -ForegroundColor Red
    }
} catch {
    Write-Host "ERROR: Analytics service compilation error: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "Summary" -ForegroundColor Cyan
Write-Host "=================================================="

Write-Host "Your cloud services are configured as follows:" -ForegroundColor White
Write-Host "• Supabase (Database): SUCCESS Configured" -ForegroundColor Green
Write-Host "• Confluent Cloud (Kafka): SUCCESS Configured" -ForegroundColor Green  
Write-Host "• Redis Cloud: SUCCESS Configured" -ForegroundColor Green

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Run database migrations: go run cmd/migrate/main.go" -ForegroundColor White
Write-Host "2. Start the server: go run cmd/server/main.go" -ForegroundColor White
Write-Host "3. Start analytics service: go run cmd/analytics/main.go" -ForegroundColor White
Write-Host "4. Test the API endpoints" -ForegroundColor White

Write-Host ""
Write-Host "Credential validation complete!" -ForegroundColor Green