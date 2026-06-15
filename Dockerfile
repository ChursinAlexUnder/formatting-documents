FROM golang:1.23-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY database ./database
COPY internal ./internal
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/formatting-documents ./cmd/main.go

FROM python:3.12-slim-bookworm AS runtime

ENV APP_ROOT=/app \
    APP_BUFFER_DIR=/app/buffer \
    APP_DATA_FILE=/app/data.json \
    PYTHON_BIN=python3 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

RUN pip install --no-cache-dir python-docx requests

COPY --from=builder /out/formatting-documents /app/formatting-documents
COPY scripts /app/scripts
COPY web /app/web
COPY data.json /app/data.json

RUN mkdir -p /app/buffer

EXPOSE 8080

CMD ["/app/formatting-documents"]
