package main

import (
	"fmt"
	"net/http"
	"text/template"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	tmp, err := template.ParseFiles("html/index.html")
	if err != nil {
		fmt.Fprintf(w, "Error!")
	}
	tmp.Execute(w, nil)
}

func handleRequest() {
	// подключение css
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))

	// Отображение страниц
	http.HandleFunc("/", homePage)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handleRequest()
}
