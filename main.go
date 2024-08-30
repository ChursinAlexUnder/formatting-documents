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
	// обработка css
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))

	// обработка js
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))

	// обработка изображений
	http.Handle("/pictures/", http.StripPrefix("/pictures/", http.FileServer(http.Dir("./pictures/"))))

	// Отображение страниц
	http.HandleFunc("/", homePage)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handleRequest()
}
