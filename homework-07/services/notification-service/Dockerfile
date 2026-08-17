FROM golang:1.26-bookworm AS builder

WORKDIR /app
COPY internal/platform/go.* ./internal/platform/
COPY services/notification-service/go.* ./services/notification-service/
WORKDIR /app/services/notification-service
RUN go mod download
WORKDIR /app
COPY internal/platform ./internal/platform
COPY services/notification-service ./services/notification-service
WORKDIR /app/services/notification-service
RUN go build -v -o /app/server ./cmd/api

FROM debian:bookworm-slim
COPY --from=builder /app/server /app/server
EXPOSE 8003
ENTRYPOINT ["/app/server"]
