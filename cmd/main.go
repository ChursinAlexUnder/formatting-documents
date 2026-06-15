package main

import (
	"formatting-documents/database"
	"formatting-documents/internal/config"
	"formatting-documents/pkg"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := config.EnsureRuntimeDirs(); err != nil {
		log.Fatalf("Не удалось подготовить рабочие каталоги: %v", err)
	}
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		dbConnStr = "user=postgres password=postgres dbname=formatting_documents host=localhost port=5432 sslmode=disable"
	}

	err := database.InitDB(dbConnStr)
	if err != nil {
		log.Printf("Предупреждение при инициализации базы данных: %v", err)
	}

	pkg.ConnectionStatic()
	pkg.HandlerPages()
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Printf("Сервер не запустился: %v", err)
	}
}
