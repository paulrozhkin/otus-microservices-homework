# Выполняем сборку в golang:1.26-bookworm
FROM golang:1.26-bookworm AS builder

# Копируем исходники и собираем сервер
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN go build -v -o server

# Запуск приложения выполняем в debian:bookworm-slim
FROM debian:bookworm-slim

# Копируем бинарник и ставим entrypoint /app/server
COPY --from=builder /app/server /app/server
EXPOSE 8000
ENTRYPOINT  ["/app/server"]