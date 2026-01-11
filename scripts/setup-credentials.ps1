# Connect 4 Multiplayer - Credential Setup Script (PowerShell)
# This script helps you set up and validate your cloud service credentials

param(
    [string]$Action = "menu"
)

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Function to check if command exists
function Test-Command {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# Function to validate environment file
function Test-EnvFile {
    param([string]$EnvFile)
    
    if (-not (Test-Path $EnvFile)) {
        Write-Error "Environment file $EnvFile not found"
        return $false
    }
    
    Write-Status "Validating environment file: $EnvFile"
    
    # Read environment variables from file
    $envVars = @{}
    Get-Content $EnvFile | ForEach-Object {
        if ($_ -match '^([^#][^=]+)=(.*)$') {
            $envVars[$matches[1]] = $matches[2]
        }
    }
    
    # Check required variables
    $requiredVars = @(
        "ENVIRONMENT",
        "DATABASE_URL",
        "KAFKA_BOOTSTRAP_SERVERS"
    )
    
    $missingVars = @()
    
    foreach ($var in $requiredVars) {
        if (-not $envVars.ContainsKey($var) -or [string]::IsNullOrEmpty($envVars[$var])) {
            $missingVars += $var
        }
    }
    
    if ($missingVars.Count -gt 0) {
        Write-Error "Missing required environment variables:"
        foreach ($var in $missingVars) {
            Write-Host "  - $var"
        }
        return $false
    }
    
    Write-Success "Environment file validation passed"
    return $true
}

# Function to test database connection
function Test-DatabaseConnection {
    Write-Status "Testing database connection..."
    
    # Load environment variables
    $envVars = @{}
    if (Test-Path ".env") {
        Get-Content ".env" | ForEach-Object {
            if ($_ -match '^([^#][^=]+)=(.*)$') {
                $envVars[$matches[1]] = $matches[2]
            }
        }
    }
    
    if (-not $envVars.ContainsKey("DATABASE_URL") -or [string]::IsNullOrEmpty($envVars["DATABASE_URL"])) {
        Write-Error "DATABASE_URL not set"
        return $false
    }
    
    $databaseUrl = $envVars["DATABASE_URL"]
    
    # Check if psql is available
    if (-not (Test-Command "psql")) {
        Write-Warning "psql not found. Install PostgreSQL client to test database connection"
        Write-Status "You can install it from: https://www.postgresql.org/download/"
        return $true
    }
    
    # Test connection with a simple query
    try {
        $result = & psql $databaseUrl -c "SELECT version();" 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Database connection successful"
            
            # Get database info
            $version = & psql $databaseUrl -t -c "SELECT version();" 2>$null | Select-Object -First 1
            if ($version) {
                Write-Status "Database version: $($version.Trim())"
            }
            
            return $true
        }
        else {
            throw "Connection failed"
        }
    }
    catch {
        Write-Error "Database connection failed"
        Write-Status "Please check your DATABASE_URL and ensure:"
        Write-Host "  1. The database server is running"
        Write-Host "  2. The credentials are correct"
        Write-Host "  3. The database exists"
        Write-Host "  4. Network connectivity is available"
        return $false
    }
}

# Function to test Kafka connection
function Test-KafkaConnection {
    Write-Status "Testing Kafka connection..."
    
    # Load environment variables
    $envVars = @{}
    if (Test-Path ".env") {
        Get-Content ".env" | ForEach-Object {
            if ($_ -match '^([^#][^=]+)=(.*)$') {
                $envVars[$matches[1]] = $matches[2]
            }
        }
    }
    
    if (-not $envVars.ContainsKey("KAFKA_BOOTSTRAP_SERVERS") -or [string]::IsNullOrEmpty($envVars["KAFKA_BOOTSTRAP_SERVERS"])) {
        Write-Error "KAFKA_BOOTSTRAP_SERVERS not set"
        return $false
    }
    
    $bootstrapServers = $envVars["KAFKA_BOOTSTRAP_SERVERS"]
    
    # For Confluent Cloud, we need API key and secret
    if ($bootstrapServers -like "*confluent.cloud*") {
        if (-not $envVars.ContainsKey("KAFKA_API_KEY") -or [string]::IsNullOrEmpty($envVars["KAFKA_API_KEY"]) -or
            -not $envVars.ContainsKey("KAFKA_API_SECRET") -or [string]::IsNullOrEmpty($envVars["KAFKA_API_SECRET"])) {
            Write-Error "Confluent Cloud requires KAFKA_API_KEY and KAFKA_API_SECRET"
            return $false
        }
        
        Write-Status "Confluent Cloud configuration detected"
        Write-Status "Bootstrap servers: $bootstrapServers"
        Write-Status "API Key: $($envVars['KAFKA_API_KEY'].Substring(0, [Math]::Min(8, $envVars['KAFKA_API_KEY'].Length)))..."
        
        # Check if confluent CLI is available
        if (Test-Command "confluent") {
            Write-Status "Testing with Confluent CLI..."
            Write-Warning "Manual verification recommended via Confluent Cloud dashboard"
        }
        else {
            Write-Warning "Confluent CLI not found. Install it for advanced testing:"
            Write-Host "  Download from: https://docs.confluent.io/confluent-cli/current/install.html"
        }
    }
    else {
        Write-Status "Local Kafka configuration detected"
        Write-Status "Bootstrap servers: $bootstrapServers"
        
        Write-Warning "Kafka connection testing requires Kafka tools"
        Write-Status "Ensure Kafka is running: docker-compose up kafka"
    }
    
    return $true
}

# Function to run database migrations
function Invoke-Migrations {
    Write-Status "Running database migrations..."
    
    if (-not (Test-Path "cmd/migrate/main.go")) {
        Write-Error "Migration tool not found at cmd/migrate/main.go"
        return $false
    }
    
    try {
        & go run cmd/migrate/main.go
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Database migrations completed successfully"
            return $true
        }
        else {
            throw "Migration failed"
        }
    }
    catch {
        Write-Error "Database migrations failed"
        return $false
    }
}

# Function to create .env file from template
function New-EnvFile {
    if (Test-Path ".env") {
        Write-Warning ".env file already exists"
        $response = Read-Host "Do you want to overwrite it? (y/N)"
        if ($response -notmatch '^[Yy]$') {
            Write-Status "Keeping existing .env file"
            return $true
        }
    }
    
    if (Test-Path ".env.example") {
        Copy-Item ".env.example" ".env"
        Write-Success "Created .env file from .env.example"
        Write-Status "Please edit .env file with your actual credentials"
        Write-Status "See docs/cloud-setup-guide.md for detailed instructions"
        return $true
    }
    else {
        Write-Error ".env.example file not found"
        return $false
    }
}

# Function to show setup instructions
function Show-SetupInstructions {
    Write-Host ""
    Write-Status "=== Connect 4 Multiplayer Credential Setup ==="
    Write-Host ""
    Write-Status "This script helps you set up and validate cloud service credentials."
    Write-Host ""
    Write-Status "Prerequisites:"
    Write-Host "  1. Supabase account and project"
    Write-Host "  2. Confluent Cloud account and cluster (optional)"
    Write-Host "  3. .env file with your credentials"
    Write-Host ""
    Write-Status "For detailed setup instructions, see:"
    Write-Host "  docs/cloud-setup-guide.md"
    Write-Host ""
}

# Function to show menu
function Show-Menu {
    Write-Host ""
    Write-Status "What would you like to do?"
    Write-Host "  1) Create .env file from template"
    Write-Host "  2) Validate environment configuration"
    Write-Host "  3) Test database connection"
    Write-Host "  4) Test Kafka connection"
    Write-Host "  5) Run database migrations"
    Write-Host "  6) Run all tests"
    Write-Host "  7) Show setup instructions"
    Write-Host "  8) Exit"
    Write-Host ""
}

# Main function
function Main {
    Show-SetupInstructions
    
    while ($true) {
        Show-Menu
        $choice = Read-Host "Enter your choice (1-8)"
        
        switch ($choice) {
            "1" {
                New-EnvFile | Out-Null
            }
            "2" {
                if (Test-EnvFile ".env") {
                    Write-Success "Environment configuration is valid"
                }
                else {
                    Write-Error "Environment configuration validation failed"
                }
            }
            "3" {
                if (Test-EnvFile ".env") {
                    Test-DatabaseConnection | Out-Null
                }
            }
            "4" {
                if (Test-EnvFile ".env") {
                    Test-KafkaConnection | Out-Null
                }
            }
            "5" {
                if (Test-EnvFile ".env") {
                    Invoke-Migrations | Out-Null
                }
            }
            "6" {
                if (Test-EnvFile ".env") {
                    Test-DatabaseConnection | Out-Null
                    Test-KafkaConnection | Out-Null
                    Invoke-Migrations | Out-Null
                    Write-Success "All tests completed"
                }
            }
            "7" {
                Show-SetupInstructions
            }
            "8" {
                Write-Status "Goodbye!"
                return
            }
            default {
                Write-Error "Invalid choice. Please enter 1-8."
            }
        }
        
        Write-Host ""
        Read-Host "Press Enter to continue..." | Out-Null
    }
}

# Handle command line arguments
switch ($Action.ToLower()) {
    "menu" { Main }
    "validate" { Test-EnvFile ".env" }
    "test-db" { Test-DatabaseConnection }
    "test-kafka" { Test-KafkaConnection }
    "migrate" { Invoke-Migrations }
    "create-env" { New-EnvFile }
    default { 
        Write-Error "Unknown action: $Action"
        Write-Status "Available actions: menu, validate, test-db, test-kafka, migrate, create-env"
    }
}