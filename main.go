package main

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/ChursinAlexUnder/Golang-website/golang/filework"
)

// TODO: увеличить шрифт для ввода текста и для заголовка на мобиле
// TODO: позже глянуть и использовать ExecuteTemplate для нескольких шаблонов вместо Execute
// TODO: сделать проверку, что файл расширения docx
// TODO: подключить и поработать с пакетом для docx (в браузере в закладках) (видимо, буду использовать питон)

func homePage(w http.ResponseWriter, r *http.Request) {
	tmp, err := template.ParseFiles("html/index.html")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
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
	http.HandleFunc("/formData", filework.FormSend)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handleRequest()
}
