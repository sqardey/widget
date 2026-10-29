# Этап 1: Сборка бинарника
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Кэшируем зависимости
COPY go.mod ./
# Если есть go.sum, раскомментируй строку ниже:
# COPY go.sum ./
RUN go mod download

COPY . .

# Статическая сборка бинарника без CGO
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/widget-server ./cmd/server/main.go

# Этап 2: Финальный легковесный образ
FROM alpine:3.21

WORKDIR /app

# Устанавливаем сертификаты доверия и базу данных часовых поясов
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/widget-server /app/widget-server

EXPOSE 8080

CMD ["/app/widget-server"]