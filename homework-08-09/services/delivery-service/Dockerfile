FROM golang:1.26-bookworm AS builder

WORKDIR /app
COPY internal/platform/go.* ./internal/platform/
COPY services/delivery-service/go.* ./services/delivery-service/
WORKDIR /app/services/delivery-service
RUN go mod download
WORKDIR /app
COPY internal/platform ./internal/platform
COPY services/delivery-service ./services/delivery-service
WORKDIR /app/services/delivery-service
RUN go build -v -o /app/server ./cmd/api

FROM debian:bookworm-slim
COPY --from=builder /app/server /app/server
EXPOSE 8005
ENTRYPOINT ["/app/server"]
