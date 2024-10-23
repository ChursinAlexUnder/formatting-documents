# Используем базовый образ Debian и добавим сюда Go и Python
FROM debian:bullseye-slim

# Установите необходимые пакеты
RUN apt-get update && apt-get install -y \
    golang \
    python3 \
    python3-pip && \
    ln -s /usr/bin/python3 /usr/bin/python && \
    apt-get clean

# Установите библиотеку python-docx через pip
RUN pip install python-docx

# Установите рабочую директорию
WORKDIR /formatting-documents

# Копируем go.mod для загрузки зависимостей
COPY go.mod /formatting-documents

# Загрузите зависимости
RUN go mod download

# Скопируйте все файлы в рабочую директорию
COPY . ./formatting-documents

# Скомпилируйте Go приложение
RUN go build -o main ./cmd/main.go

# Expose port 8080 to the outside world
EXPOSE 8080

# Команда для запуска вашего Go сервера
CMD ["./main"]
