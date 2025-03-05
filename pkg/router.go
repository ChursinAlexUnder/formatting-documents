package pkg

import (
	"formatting-documents/internal/interfaces"
	"net/http"
)

func ConnectionStatic() {
	// обработка статических файлов
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../web/static/"))))
}

func HandlerPages() {
	// Отображение страниц
	http.HandleFunc("/", interfaces.MainPage)
	http.HandleFunc("/menu", interfaces.ShowOptions)
	http.HandleFunc("/download", interfaces.SendDocument)
	http.HandleFunc("/error", interfaces.ErrorPage)
	http.HandleFunc("/errortime", interfaces.ErrorTimePage)
	http.HandleFunc("/events", interfaces.SSEChannel)
}
