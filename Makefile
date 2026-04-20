.PHONY: build build-prod build-backend-prod build-frontend-prod docker-build install frontend clean check-ports

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

# Build Go backend for production
build-backend-prod:
	@mkdir -p dist
	CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o dist/wsapi .
	@echo "Backend production build complete"

# Build Next.js frontend for production
build-frontend-prod:
	cd frontend && npm ci && npm run build
	@echo "Frontend production build complete"

# Full production build (backend + frontend artifacts)
build-prod: build-backend-prod build-frontend-prod
	@echo "Production build complete"

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

# Validate production ports
check-ports:
	bash docker/check-ports.sh

# Build Docker images for production
docker-build:
	docker compose build

# Clean build artifacts
clean:
	rm -f wsapi
	rm -rf dist
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
	@echo "  build-backend-prod - Production Go build"
	@echo "  build-frontend-prod - Production Next.js build"
	@echo "  build-prod - Production backend + frontend build"
	@echo "  docker-build - Build Docker images"
	@echo "  check-ports - Verify occupied ports"
	@echo "  static    - Copy frontend to static folder"
	@echo "  build     - Build frontend, static copy, and backend"
	@echo "  dev       - Run frontend in development mode"
	@echo "  clean     - Clean build artifacts"
	@echo "  test      - Run tests"
	@echo "  lint      - Run linter"
