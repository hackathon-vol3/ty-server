.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run -v ./...

.PHONY: up
up:
	docker compose up -d --build

.PHONY: down
down:
	docker compose down -v
