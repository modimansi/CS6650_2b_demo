# PowerShell build script for Lambda function

Write-Host "=========================================================" -ForegroundColor Cyan
Write-Host "     Building Lambda Payment Processor                  " -ForegroundColor Cyan
Write-Host "=========================================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: Go is not installed" -ForegroundColor Red
    exit 1
}

Write-Host "Installing dependencies..." -ForegroundColor Yellow
go mod download

Write-Host "Building for Linux (Lambda environment)..." -ForegroundColor Yellow
# Set environment variables for cross-compilation
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

# Build the function
go build -tags lambda.norpc -o bootstrap main.go

Write-Host "Creating deployment package..." -ForegroundColor Yellow
# Remove old zip if exists
if (Test-Path function.zip) {
    Remove-Item function.zip
}

# Create zip file
Compress-Archive -Path bootstrap -DestinationPath function.zip -Force

Write-Host ""
Write-Host "SUCCESS: Lambda function built successfully!" -ForegroundColor Green
Write-Host ""
Get-Item function.zip | Format-Table Name, Length
Write-Host ""
Write-Host "Deploy with: cd ../../terraform && terraform apply" -ForegroundColor Yellow

