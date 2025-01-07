# Используем официальный образ Golang
FROM golang:1.20

# Устанавливаем рабочую директорию
WORKDIR /order

# Копируем файлы go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Запускаем тесты (если тесты не нужны, удалите эту строку)
RUN go test ./...

# Компилируем приложение
RUN go build -o order main.go

# Открываем порт для приложения
EXPOSE 5002

# Запускаем приложение
CMD ["./order"]
