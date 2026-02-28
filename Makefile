.PHONY: all build run test clean deps fmt lint itest cover migrate-up migrate-docker seed seed-docker swagger up down trivy verify test-docker itest-docker help

# --- Dependencies & Setup ---

# Install Go dependencies
deps:
	go mod download

# Format code using go fmt
fmt:
	go fmt ./...

# Lint code using golangci-lint
lint:
	golangci-lint run

PWD := $(shell pwd)

# --- Testing ---

# Run unit tests locally (fast, no DB required)
test-unit-local:
	go test -v -coverpkg=./internal/... ./tests/unit/...

# Run unit tests inside Docker (consistent environment)
test-unit:
	docker run --rm -v $(CURDIR):/app -w /app golang:1.23-alpine go test -v -coverpkg=./internal/... ./tests/unit/...

# Run integration tests locally (requires running DB via 'make up')
test-integration-local:
	set DB_URL=postgres://admin:secret@localhost:5433/autorepair?sslmode=disable&& go test -v -tags=integration -coverpkg=./internal/... ./tests/integration/...

# Run integration tests inside Docker (connects to Docker DB)
test-integration:
	docker run --rm --network go_autorepair-net -v $(CURDIR):/app -w /app -e DB_URL="postgres://admin:secret@db:5432/autorepair?sslmode=disable" golang:1.23-alpine go test -v -tags=integration -coverpkg=./internal/... ./tests/integration/...

# Run ALL tests (Unit + Integration) and generate coverage report locally
test-cover-local:
	set DB_URL=postgres://admin:secret@localhost:5433/autorepair?sslmode=disable&& go test -v -tags=integration -coverpkg=./internal/... -coverprofile=coverage.out ./tests/...
	grep -v -E "docs/|cmd/|seeds/" coverage.out > coverage_filtered.out
	go tool cover -func=coverage_filtered.out

# Generate HTML coverage report locally
test-cover-html: test-cover-local
	go tool cover -html=coverage_filtered.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Run ALL tests and generate coverage report inside Docker (Recommended for CI/Consistency)
test-cover:
	docker run --rm --network go_autorepair-net -v $(CURDIR):/app -w /app -e DB_URL="postgres://admin:secret@db:5432/autorepair?sslmode=disable" golang:1.23-alpine sh -c "go test -v -coverpkg=./internal/... -coverprofile=coverage.out -tags=integration ./tests/... && grep -v -E 'docs/|cmd/|seeds/' coverage.out > coverage_filtered.out && go tool cover -func=coverage_filtered.out"

# --- Database & Migrations ---

# Run migrations locally
migrate-up:
	migrate -path migrations -database "postgres://admin:secret@localhost:5433/autorepair?sslmode=disable" up

# Run migrations inside Docker
migrate-docker:
	docker run --rm --network go_autorepair-net -v $(CURDIR)/migrations:/migrations migrate/migrate -path=/migrations/ -database "postgres://admin:secret@db:5432/autorepair?sslmode=disable" up

# Seed database locally
seed:
	go run seeds/main.go

# Seed database inside Docker
seed-docker:
	docker compose run --rm app go run seeds/main.go

# --- Documentation ---

# Generate Swagger docs
swagger:
	swag init -g cmd/api/main.go

# --- Docker Control ---

# Start environment (App + DB + Sonar)
up:
	docker-compose up -d --build

# Stop environment and remove volumes
down:
	docker-compose down -v

# Access container shell
login:
	docker compose exec app sh

# Watch logs
watch-logs:
	docker compose logs -f app

# --- Security & Verification ---

# Run Trivy security scan
trivy:
	trivy fs .

# Run full verification suite (fmt, lint, test, security)
verify: fmt lint test itest swagger cover trivy

# Show help
help:
	@echo "Available commands:"
	@echo "  make up              - Start Docker environment"
	@echo "  make down            - Stop Docker environment"
	@echo "  make test-unit       - Run Unit Tests (Docker)"
	@echo "  make test-integration- Run Integration Tests (Docker)"
	@echo "  make test-cover      - Run All Tests & Coverage (Docker)"
	@echo "  make migrate-docker  - Run DB Migrations (Docker)"
	@echo "  make seed-docker     - Seed DB with dummy data (Docker)"
