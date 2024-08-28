package main

import (
	"fmt"
	"net/http"
	"text/template"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	tmp, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Fprintf(w, "Error!")
	}
	tmp.Execute(w, nil)
}

func informationPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "It's information page, oooh!")
}

func handleRequest() {
	// подключение css
	http.Handle("/style/", http.StripPrefix("/style/", http.FileServer(http.Dir("./style/"))))

	// Отображение страниц
	http.HandleFunc("/", homePage)
	http.HandleFunc("/information/", informationPage)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handleRequest()
}
