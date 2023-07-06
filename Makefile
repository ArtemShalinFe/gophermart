# Makefile

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
            -gophermart-port=8078 \
            -gophermart-database-uri="postgresql://postgres:postgres@postgres/praktikum?sslmode=disable" \
            -accrual-binary-path=cmd/accrual/accrual_darwin_arm64 \
            -accrual-host=localhost \
            -accrual-port=8080 \
            -accrual-database-uri="postgresql://postgres:postgres@postgres/praktikum?sslmode=disable"

.PHONY: lint
lint:
	docker run --rm \
    -v $(shell pwd):/app \
    -v $(shell pwd)/golangci-lint/.cache/golangci-lint/v1.53.3:/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.53.3 \
        golangci-lint run \
        -c .golangci-lint.yml \
    > ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json | rm ./golangci-lint/report-unformatted.json

# POSTGRESQL

.PHONY: pg
run-pg:
	docker run --rm \
		--name=postgresql \
		-v $(abspath ./db/init/):/docker-entrypoint-initdb.d \
		-v $(abspath ./db/data/):/var/lib/postgresql/data \
		-e POSTGRES_PASSWORD="gopher" \
		-d \
		-p 5432:5432 \
		postgres:15.3

.PHONY: stop-pg
stop-pg:
	docker stop postgresql

.PHONY: clean-data
clean-data:
	sudo rm -rf ./db/data/