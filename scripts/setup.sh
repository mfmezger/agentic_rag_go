#!/bin/bash
# Setup script for Agentic RAG Go

set -e

echo "🚀 Setting up Agentic RAG Go..."

# Check for required tools
command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed."; exit 1; }

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy

# Install dev tools
echo "🔧 Installing development tools..."
go install mvdan.cc/gofumpt@latest
go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest

# Create config from example
if [ ! -f configs/config.yaml ]; then
    echo "📝 Creating config file from example..."
    cp configs/config.example.yaml configs/config.yaml
fi

echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Set your GOOGLE_API_KEY environment variable"
echo "  2. Update configs/config.yaml with your settings"
echo "  3. Run 'make run' to start the application"
