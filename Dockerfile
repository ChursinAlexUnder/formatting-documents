# Используем базовый образ с Python и Go
FROM golang:1.21 as builder
COPY . /www
WORKDIR /www
RUN go mod download
RUN go build -o main .

FROM python:3.12.0-slim

WORKDIR /www

# Скопируем исполняемый файл Go из предыдущего шага
COPY --from=builder /www/main .

# Скопируем все необходимые файлы для Python сервера
COPY . .

# Установим Python зависимости
RUN pip install --no-cache-dir -r requirements.txt

# Команда запуска обоих процессов
CMD ["sh", "-c", "python /www/python/editdocument.py & ./main"]