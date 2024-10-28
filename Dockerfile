# Используем базовый образ Debian и добавим сюда Go и Python
# FROM debian:bullseye-slim

# Используем базовый образ с Go версии 1.23
FROM golang:1.23-bullseye

# Установите необходимые пакеты
RUN apt-get update && apt-get install -y \
    golang \
    python3 \
    apt-utils && \
    python3-pip && \
    ln -s /usr/bin/python3 /usr/bin/python && \
    apt-get clean

# Установите библиотеку python-docx через pip
RUN pip install python-docx

# Установите рабочую директорию
WORKDIR /

# Копируем go.mod для загрузки зависимостей
COPY go.mod /

# Скопируйте все файлы в рабочую директорию
COPY . /

# Скомпилируйте Go приложение
RUN go build -o main ./cmd/main.go

# Expose port 8080 to the outside world
EXPOSE 8080

# Команда для запуска вашего Go сервера
CMD ["./main"]
