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
    echo "⚠️  Python3 not found."
    if command -v brew &> /dev/null; then
        brew install python3
    elif command -v apt-get &> /dev/null; then
        echo "Run: sudo apt-get install python3 python3-pip"
        exit 1
    elif command -v dnf &> /dev/null; then
        echo "Run: sudo dnf install python3 python3-pip"
        exit 1
    else
        echo "❌ Install Python3 from https://www.python.org/downloads/"
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

# Upgrade pip/setuptools (pip installs happen inside update-tools; this just primes pip itself)
echo
echo "Upgrading pip..."
pip3 install --upgrade pip setuptools wheel 2>/dev/null \
    || pip3 install --break-system-packages --upgrade pip setuptools wheel 2>/dev/null \
    || true

# Check for Rust (needed for findomain)
if ! command -v cargo &> /dev/null; then
    echo
    echo "⚠️  Rust/cargo not found (needed for findomain)"
    echo "Install: https://rustup.rs/"
    echo "Note: update-tools will attempt a GitHub release download as fallback."
fi

# Build binary (CGO_ENABLED=0 produces a static binary that works on all Linux variants
# including ARM64 containers where the CGO cross-compiler may lack -m64 support)
echo
echo "Building ratnosint7..."
CGO_ENABLED=0 go build -o ratnosint7 ./cmd/ratnosint7
echo "✓ Built: ./ratnosint7"

# Install enumeration tools
echo
echo "Installing enumeration tools..."
./ratnosint7 update-tools

echo
echo "✓ Setup complete. Run: ./ratnosint7 scan"
