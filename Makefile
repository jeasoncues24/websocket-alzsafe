.PHONY: build install frontend clean

# Default target
all: build

# Install frontend dependencies
install:
	cd frontend && npm install

# Build frontend
frontend:
	cd frontend && npm run build
	@echo "Frontend built successfully"

# Build Go backend
backend:
	go build -o wsapi .

# Copy frontend to static folder
static:
	rm -rf internal/http/static
	cp -r frontend/out internal/http/static

# Full build (frontend + backend)
build: frontend static backend
	@echo "Full build complete"

# Development mode
dev:
	cd frontend && npm run dev

# Clean build artifacts
clean:
	rm -f wsapi
	rm -rf frontend/out
	rm -rf internal/http/static

# Run tests
test:
	go test ./...

# Run linter
lint:
	go vet ./...

# Help
help:
	@echo "Available targets:"
	@echo "  install    - Install frontend dependencies"
	@echo "  frontend  - Build Next.js frontend"
	@echo "  backend   - Build Go backend"
	@echo "  static    - Copy frontend to static folder"
	@echo "  build     - Build frontend, static copy, and backend"
	@echo "  dev       - Run frontend in development mode"
	@echo "  clean     - Clean build artifacts"
	@echo "  test      - Run tests"
	@echo "  lint      - Run linter"