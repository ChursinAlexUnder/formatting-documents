package filework

import (
	"fmt"
	"net/http"
	"text/template"
)

type Answer struct {
	FileName, Change string
}

func FormSend(w http.ResponseWriter, r *http.Request) {
	// получение значения полей
	var change string
	file, fileData, err := r.FormFile("document-file")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	defer file.Close()
	change = r.FormValue("change")

	// загрузка и распарсивание HTML-шаблона из файла temp.html для дальнейшей работы
	tmp, err := template.ParseFiles("./html/temp.html")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	// добавление данных на страницу
	answer := Answer{fileData.Filename, change}
	tmp.Execute(w, answer)
}
