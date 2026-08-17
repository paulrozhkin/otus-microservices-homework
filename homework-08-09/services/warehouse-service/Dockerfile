FROM golang:1.26-bookworm AS builder

WORKDIR /app
COPY internal/platform/go.* ./internal/platform/
COPY services/warehouse-service/go.* ./services/warehouse-service/
WORKDIR /app/services/warehouse-service
RUN go mod download
WORKDIR /app
COPY internal/platform ./internal/platform
COPY services/warehouse-service ./services/warehouse-service
WORKDIR /app/services/warehouse-service
RUN go build -v -o /app/server ./cmd/api

FROM debian:bookworm-slim
COPY --from=builder /app/server /app/server
EXPOSE 8004
ENTRYPOINT ["/app/server"]
