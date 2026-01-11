#!/usr/bin/env pwsh
# PowerShell script to validate Kafka Cloud setup

Write-Host "🔍 Kafka Cloud Validation Script" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Host "❌ .env file not found!" -ForegroundColor Red
    exit 1
}

# Load environment variables
Get-Content ".env" | ForEach-Object {
    if ($_ -match "^([^#=]+)=(.*)$") {
        $name = $matches[1].Trim()
        $value = $matches[2].Trim()
        [Environment]::SetEnvironmentVariable($name, $value, "Process")
    }
}

# Check Kafka configuration
Write-Host "📋 Checking Kafka Configuration..." -ForegroundColor Yellow

$kafkaBootstrap = [Environment]::GetEnvironmentVariable("KAFKA_BOOTSTRAP_SERVERS")
$kafkaApiKey = [Environment]::GetEnvironmentVariable("KAFKA_API_KEY")
$kafkaApiSecret = [Environment]::GetEnvironmentVariable("KAFKA_API_SECRET")
$kafkaTopic = [Environment]::GetEnvironmentVariable("KAFKA_TOPIC")
$kafkaGroup = [Environment]::GetEnvironmentVariable("KAFKA_CONSUMER_GROUP")

$allGood = $true

if ([string]::IsNullOrEmpty($kafkaBootstrap)) {
    Write-Host "   ❌ KAFKA_BOOTSTRAP_SERVERS not set" -ForegroundColor Red
    $allGood = $false
} else {
    Write-Host "   ✅ Bootstrap Servers: $kafkaBootstrap" -ForegroundColor Green
}

if ([string]::IsNullOrEmpty($kafkaApiKey)) {
    Write-Host "   ❌ KAFKA_API_KEY not set" -ForegroundColor Red
    $allGood = $false
} else {
    Write-Host "   ✅ API Key: $($kafkaApiKey.Substring(0, 8))..." -ForegroundColor Green
}

if ([string]::IsNullOrEmpty($kafkaApiSecret)) {
    Write-Host "   ❌ KAFKA_API_SECRET not set" -ForegroundColor Red
    $allGood = $false
} else {
    Write-Host "   ✅ API Secret: $($kafkaApiSecret.Substring(0, 8))..." -ForegroundColor Green
}

if ([string]::IsNullOrEmpty($kafkaTopic)) {
    Write-Host "   ❌ KAFKA_TOPIC not set" -ForegroundColor Red
    $allGood = $false
} else {
    Write-Host "   ✅ Topic: $kafkaTopic" -ForegroundColor Green
}

if ([string]::IsNullOrEmpty($kafkaGroup)) {
    Write-Host "   ❌ KAFKA_CONSUMER_GROUP not set" -ForegroundColor Red
    $allGood = $false
} else {
    Write-Host "   ✅ Consumer Group: $kafkaGroup" -ForegroundColor Green
}

Write-Host ""

if (-not $allGood) {
    Write-Host "❌ Configuration issues found. Please check your .env file." -ForegroundColor Red
    exit 1
}

# Test compilation
Write-Host "🔧 Testing Compilation..." -ForegroundColor Yellow

Write-Host "   Testing producer compilation..." -ForegroundColor Gray
$producerTest = & go build -o temp_producer.exe scripts/test-kafka-cloud.go 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ Producer test compiles" -ForegroundColor Green
    Remove-Item "temp_producer.exe" -ErrorAction SilentlyContinue
} else {
    Write-Host "   ❌ Producer compilation failed: $producerTest" -ForegroundColor Red
    $allGood = $false
}

Write-Host "   Testing consumer compilation..." -ForegroundColor Gray
$consumerTest = & go build -o temp_consumer.exe scripts/test-kafka-consumer.go 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ Consumer test compiles" -ForegroundColor Green
    Remove-Item "temp_consumer.exe" -ErrorAction SilentlyContinue
} else {
    Write-Host "   ❌ Consumer compilation failed: $consumerTest" -ForegroundColor Red
    $allGood = $false
}

Write-Host ""

# Summary and next steps
if ($allGood) {
    Write-Host "🎉 Kafka Cloud Setup Validation: SUCCESS!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📋 Your Kafka configuration is ready:" -ForegroundColor White
    Write-Host "   • Confluent Cloud credentials configured" -ForegroundColor Green
    Write-Host "   • Test scripts compile successfully" -ForegroundColor Green
    Write-Host "   • Ready for message production and consumption" -ForegroundColor Green
    Write-Host ""
    Write-Host "🚀 Next Steps - Test Your Kafka Connection:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "1. Test Producer (send messages):" -ForegroundColor White
    Write-Host "   go run scripts/test-kafka-cloud.go" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "2. Test Consumer (receive messages):" -ForegroundColor White
    Write-Host "   go run scripts/test-kafka-consumer.go" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "3. Check Confluent Cloud Console:" -ForegroundColor White
    Write-Host "   https://confluent.cloud/environments" -ForegroundColor Cyan
    Write-Host "   → Your Cluster → Topics → $kafkaTopic" -ForegroundColor Gray
    Write-Host ""
    Write-Host "4. Start Analytics Service:" -ForegroundColor White
    Write-Host "   go run cmd/analytics/main.go" -ForegroundColor Cyan
    Write-Host ""
} else {
    Write-Host "❌ Kafka Cloud Setup Issues Found" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please fix the issues above and run this script again." -ForegroundColor Yellow
    Write-Host "Check your .env file and Confluent Cloud credentials." -ForegroundColor Yellow
}

Write-Host ""