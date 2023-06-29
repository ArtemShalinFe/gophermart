# Makefile

.PHONY: all
all: ;

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

