# Используем базовый образ Debian и добавим сюда Go и Python
FROM debian:bullseye-slim

# Установить необходимые пакеты
RUN apt-get update && apt-get install -y \
    golang \
    python3 \
    python3-pip && \
    ln -s /usr/bin/python3 /usr/bin/python && \
    apt-get clean

# Установить библиотеку python-docx через pip
RUN pip install python-docx

# Установить рабочую директорию
WORKDIR /

# Копировать go.mod для загрузки зависимостей
COPY go.mod go.sum /

# Загрузить зависимости
RUN go mod download

# Скопировать все файлы в рабочую директорию
COPY . /

# Скомпилировать Go приложение
RUN go build -o main ./cmd/main.go

# Expose port 8080 to the outside world
EXPOSE 8080

# Команда для запуска вашего Go сервера
CMD ["./main"]
