FROM golang:1.20 AS build

WORKDIR /usr/src/gophermart
COPY . .

RUN go mod download && go mod verify

RUN go build -C ./cmd/gophermart -o /usr/local/bin/gophermart

FROM golang:1.20-alpine

ENV RUN_ADDRESS=${RUN_ADDRESS}
ENV DATABASE_URI=${DATABASE_URI}
ENV ACCRUAL_SYSTEM_ADDRESS=${ACCRUAL_SYSTEM_ADDRESS}

COPY --from=build /usr/local/bin/gophermart /app/gophermart
RUN chmod +x /app/gophermart

CMD /app/gophermart