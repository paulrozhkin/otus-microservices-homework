FROM golang:1.26-bookworm AS builder

WORKDIR /app
COPY internal/platform/go.* ./internal/platform/
COPY services/order-service/go.* ./services/order-service/
WORKDIR /app/services/order-service
RUN go mod download
WORKDIR /app
COPY internal/platform ./internal/platform
COPY services/order-service ./services/order-service
WORKDIR /app/services/order-service
RUN go build -v -o /app/server ./cmd/api

FROM debian:bookworm-slim
COPY --from=builder /app/server /app/server
EXPOSE 8002
ENTRYPOINT ["/app/server"]
