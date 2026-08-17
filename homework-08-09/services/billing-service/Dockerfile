FROM golang:1.26-bookworm AS builder

WORKDIR /app
COPY internal/platform/go.* ./internal/platform/
COPY services/billing-service/go.* ./services/billing-service/

WORKDIR /app/services/billing-service
RUN go mod download

WORKDIR /app
COPY internal/platform ./internal/platform
COPY services/billing-service ./services/billing-service

WORKDIR /app/services/billing-service
RUN go build -v -o /app/server ./cmd/api

FROM debian:bookworm-slim

COPY --from=builder /app/server /app/server
EXPOSE 8001
ENTRYPOINT ["/app/server"]
