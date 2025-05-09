# Этап 1: Сборка Go-приложения
FROM golang:1.23-alpine AS builder

# Устанавливаем необходимые зависимости для сборки
RUN apk add --no-cache git

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod ./
RUN go mod download

# Копируем исходный код
COPY . .

# Компилируем приложение статически
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o video-downloader main.go

# Этап 2: Финальный образ на базе Alpine
FROM alpine:3.18

# Устанавливаем FFmpeg
RUN apk add --no-cache ffmpeg

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем скомпилированное приложение
COPY --from=builder /app/video-downloader .

# Указываем точку входа
ENTRYPOINT ["./video-downloader"]
