package main

import (
	"formatting-documents/database"
	"formatting-documents/pkg"
	"log"
	"net/http"
	"os"
)

func main() {
	// Initialize database
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		dbConnStr = "user=postgres password=postgres dbname=formatting_documents host=localhost port=5432 sslmode=disable"
	}

	err := database.InitDB(dbConnStr)
	if err != nil {
		log.Printf("Warning: database initialization: %v", err)
		// Don't exit, app can work without DB
	}

	pkg.ConnectionStatic()
	pkg.HandlerPages()
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Printf("Error: the server did not start: %v", err)
	}
}
