#!/bin/bash
set -e

echo "ratnosint7 — Setup & Build"
echo "=========================="
echo

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Go not found. Install: https://golang.org/dl/"
    exit 1
fi
echo "✓ Go $(go version | awk '{print $3}')"

# Check Python3
if ! command -v python3 &> /dev/null; then
    echo "⚠️  Python3 not found. Installing..."
    if command -v brew &> /dev/null; then
        brew install python3
    else
        echo "❌ Install Python3 manually: https://www.python.org/downloads/"
        exit 1
    fi
fi
echo "✓ Python3 $(python3 --version | awk '{print $2}')"

# Check pip3
if ! command -v pip3 &> /dev/null; then
    echo "⚠️  pip3 not found. Installing..."
    python3 -m ensurepip --upgrade
fi
echo "✓ pip3"

# Install Python tools
echo
echo "Installing Python dependencies..."
pip3 install --upgrade pip setuptools wheel 2>/dev/null || true
pip3 install sublist3r 2>/dev/null && echo "✓ sublist3r" || echo "⚠️  sublist3r failed"

# Check for Rust (needed for findomain)
if ! command -v cargo &> /dev/null; then
    echo
    echo "⚠️  Rust/cargo not found (needed for findomain)"
    echo "Install: https://rustup.rs/"
    echo "Or use: brew install findomain"
fi

# Build binary
echo
echo "Building ratnosint7..."
go build -o ratnosint7 ./cmd/ratnosint7
echo "✓ Built: ./ratnosint7"

# Install enumeration tools
echo
echo "Installing enumeration tools..."
./ratnosint7 update-tools

echo
echo "✓ Setup complete. Run: ./ratnosint7 scan"
