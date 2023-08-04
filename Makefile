# Makefile
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: all
all: ;

# BUILD
.PHONY: build
build:
	go build -C ./cmd/gophermart -o $(shell pwd)/cmd/gophermart/gophermart

# TESTS
.PHONY: tests
tests: build
	go test ./... -v -race
	gophermarttest-darwin-arm64 \
		-test.v -test.run=^TestGophermart$ \
            -gophermart-binary-path=cmd/gophermart/gophermart \
            -gophermart-host=localhost \
            -gophermart-port=8080 \
            -gophermart-database-uri="postgresql://postgres:gopher@localhost:5432/gophermart?sslmode=disable" \
            -accrual-binary-path=cmd/accrual/accrual_darwin_arm64 \
            -accrual-host=localhost \
            -accrual-port=8078 \
            -accrual-database-uri="postgresql://postgres:gopher@localhost:5432/accrual?sslmode=disable"

.PHONY: lint
lint:
	[ -d $(ROOT_DIR)/golangci-lint ] || mkdir -p $(ROOT_DIR)/golangci-lint
	docker run --rm \
    -v $(ROOT_DIR):/app \
    -v $(ROOT_DIR)/golangci-lint/.cache:/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.53.3 \
        golangci-lint run \
        -c .golangci-lint.yml \
    > ./golangci-lint/report.json

# POSTGRESQL
.PHONY: run-pg
run-pg:
	docker run --rm \
		--name=postgresql \
		-v $(ROOT_DIR)/deployments/db/init/:/docker-entrypoint-initdb.d \
		-v $(ROOT_DIR)/deployments/db/data/:/var/lib/postgresql/data \
		-e POSTGRES_PASSWORD=gopher \
		-d \
		-p 5432:5432 \
		postgres:15.3

.PHONY: stop-pg
stop-pg:
	docker stop postgresql

.PHONY: clean-data
clean-data:
	rm -rf ./deployments/db/data/	

# MOCKS
.PHONY: mocks
mocks:
	mockgen -source=internal/server/handlers.go -destination=internal/server/mock_handlers.go -package server

.PHONY: clean-mocks
clean-mocks:
	rm internal/server/mock_handlers.go