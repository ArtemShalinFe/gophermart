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