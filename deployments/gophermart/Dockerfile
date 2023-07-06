FROM golang:1.20.2 AS build-stage

WORKDIR /app

COPY . .

RUN go mod download && go mod verify

RUN go build -C ./cmd/gophermart -o /gophermart

FROM gcr.io/distroless/base-debian11

WORKDIR /

ENV RUN_ADDRESS=localhost:${GOPHERMART_PORT}
ENV DATABASE_URI=${DATABASE_URI}
ENV ACCRUAL_SYSTEM_ADDRESS=${ACCRUAL_SYSTEM_ADDRESS}

EXPOSE ${GOPHERMART_PORT}

COPY --from=build-stage /gophermart /gophermart

USER nonroot:nonroot

ENTRYPOINT ["/gophermart"]