FROM golang:1.24-alpine AS builder

WORKDIR /usr/local/src

# Установка необходимых утилит и библиотек для сборки.
RUN apk --no-cache add bash git make gcc gettext musl-dev

# Копирование файлов зависимостей для кэширования слоев Docker.
COPY go.mod go.sum ./
# Загрузка Go-зависимостей проекта.
RUN go mod download

# Копирование исходного кода приложения.
COPY app ./app
# Копирование файла .env с переменными окружения.
COPY .env ./

# Сборка Go-приложения в исполняемый файл 'app' в директорию /usr/local/src/bin.
RUN go build -o /usr/local/src/bin/app /usr/local/src/app/cmd/main.go

# --- Этап сборки 'runner' ---
# Создание финального легковесного образа на основе Alpine Linux.
FROM alpine AS runner

# Копирование исполняемого файла 'app' из этапа 'builder' в финальный образ.
COPY --from=builder /usr/local/src/bin/app /
# Копирование файла .env из этапа 'builder' в финальный образ.
COPY --from=builder /usr/local/src/.env /

# Определение команды для запуска приложения при старте контейнера.
CMD ["/app"]