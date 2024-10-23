package interfaces

import (
	"net/http"
)

func ConnectionStatic() {
	// обработка статических файлов
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../web/static/"))))
}

func HandlerPages() {
	// Отображение страниц
	http.HandleFunc("/", MainPage)
	http.HandleFunc("/downloadPage", SendDocumentPage)
	http.HandleFunc("/download", SendDocument)
}
