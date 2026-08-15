FROM golang:1.26-bookworm AS builder

WORKDIR /app
COPY internal/platform/go.* ./internal/platform/
COPY services/user-service/go.* ./services/user-service/

WORKDIR /app/services/user-service
RUN go mod download

WORKDIR /app
COPY internal/platform ./internal/platform
COPY services/user-service ./services/user-service

WORKDIR /app/services/user-service
RUN go build -v -o /app/server ./cmd/api

FROM debian:bookworm-slim

COPY --from=builder /app/server /app/server
EXPOSE 8000
ENTRYPOINT ["/app/server"]
