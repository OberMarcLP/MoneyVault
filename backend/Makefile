.PHONY: test test-unit test-verbose test-coverage test-integration test-integration-setup test-integration-teardown build run

test:
	go test ./internal/... -count=1

test-unit:
	go test ./internal/... -count=1 -short

test-verbose:
	go test ./internal/... -v -count=1

test-coverage:
	go test ./internal/... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML report: go tool cover -html=coverage.out"

test-integration:
	go test ./tests/... ./internal/repositories/... -p 1 -count=1 -run Integration -v

test-integration-setup:
	docker compose -f ../docker-compose.test.yml up -d --wait

test-integration-teardown:
	docker compose -f ../docker-compose.test.yml down

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server
