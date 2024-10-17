# # Use the official Go image as the base image
# FROM golang:1.21.6

# # Set the working directory in the container
# WORKDIR /www

# # Copy the application files into the working directory
# COPY . /www

# # Build the application
# RUN go build -o main .

# # Expose port 8080
# EXPOSE 8080

# # Define the entry point for the container
# CMD ["./main"]

# Используем базовый образ Debian и добавим сюда Go и Python
FROM debian:bullseye-slim

# Установите необходимые пакеты
RUN apt-get update && apt-get install -y \
    golang \
    python3 \
    python3-pip && \
    ln -s /usr/bin/python3 /usr/bin/python && \
    apt-get clean

# Установите рабочую директорию
WORKDIR /www

# Скопируйте все файлы в рабочую директорию
COPY . .

# Скомпилируйте Go приложение
RUN go build -o server main.go

# Команда для запуска вашего Go сервера
CMD ["./main"]
