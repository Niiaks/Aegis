build:
	@go build -o bin/Aegis cmd/Aegis/main.go

test:
	@go test -v ./...

run: build
	@./bin/Aegis

migration:
	@migrate create -ext sql -dir cmd/migrate/migrations $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down

# Run Workers
run-relay:
	@go run ./cmd/outbox-relay

run-webhook:
	@go run ./cmd/workers/webhook

run-balance:
	@go run ./cmd/workers/balance

# Run all workers (Note: this runs them in the background in most shells)
workers:
	@make run-relay & make run-webhook & make run-balance
