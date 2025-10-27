#!/bin/bash
set -e

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║     Building Lambda Payment Processor                     ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ ERROR: Go is not installed"
    exit 1
fi

echo "→ Installing dependencies..."
go mod download

echo "→ Building for Linux (Lambda environment)..."
# Build for Linux (Lambda runs on Amazon Linux 2)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap main.go

echo "→ Creating deployment package..."
# Remove old zip if exists
rm -f function.zip

# Create zip file
zip function.zip bootstrap

echo ""
echo "✅ Lambda function built successfully!"
echo ""
ls -lh function.zip
echo ""
echo "Deploy with: cd ../../terraform && terraform apply"

