.PHONY: help setup build clean test install-tools

help:
	@echo "ratnosint7 Makefile"
	@echo "==================="
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  setup          Setup dependencies + build + install tools (first time)"
	@echo "  build          Build binary (./ratnosint7)"
	@echo "  build-clean    Clean build (removes cache)"
	@echo "  install-tools  Install/update enumeration tools"
	@echo "  clean          Remove binary + cache"
	@echo "  run            Run scan (interactive)"
	@echo "  test           Run unit tests"
	@echo ""

setup:
	@bash setup.sh

build:
	@echo "Building ratnosint7..."
	@go build -o ratnosint7 ./cmd/ratnosint7
	@echo "✓ Built: ./ratnosint7"

build-clean:
	@echo "Clean build..."
	@go clean -cache
	@go clean
	@go build -o ratnosint7 ./cmd/ratnosint7
	@echo "✓ Built: ./ratnosint7"

install-tools:
	@./ratnosint7 update-tools

clean:
	@echo "Cleaning..."
	@rm -f ratnosint7
	@go clean -cache
	@echo "✓ Cleaned"

run:
	@./ratnosint7 scan

test:
	@go test -v ./...

.DEFAULT_GOAL := help
