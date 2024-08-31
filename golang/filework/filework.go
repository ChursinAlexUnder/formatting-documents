package filework

import (
	"fmt"
	"net/http"
	"text/template"
)

// TODO: сделать так, чтобы данные выводились именно внутрь страницы, а не поверх кода html

func FormSend(w http.ResponseWriter, r *http.Request) {
	// получение значения полей
	var change string
	file, filedata, err := r.FormFile("document-file")
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
	fmt.Fprintf(w, "Название файла: %s\n", filedata.Filename)
	fmt.Fprintf(w, "Размер файла: %d\n", filedata.Size)
	fmt.Fprintf(w, "Особые пожелания пользователя к файлу: %s\n", change)

	tmp.Execute(w, nil)
}
