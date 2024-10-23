package main

import (
	"formatting-documents/pkg"
	"log"
	"net/http"
)

func main() {
	pkg.ConnectionStatic()
	pkg.HandlerPages()
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Print("Error: the server did not start")
	} else {
		log.Print("The server has started successfully")
	}
}
