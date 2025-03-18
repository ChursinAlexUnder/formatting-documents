package interfaces

import (
	"encoding/json"
	"fmt"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/infrastructure"
	"formatting-documents/internal/services"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"text/template"
)

func MainPage(w http.ResponseWriter, r *http.Request) {
	var (
		wrongData domain.WrongData = domain.WrongData{}
		data      domain.Answer
		err       error
	)
	err = services.CheckDataJSON()
	if err != nil {
		fmt.Fprintf(w, "Error checking JSON data: %v", err)
		return
	}
	if r.Method == http.MethodPost {
		data, wrongData, err = ManagementData(w, r)
		if err != nil && err.Error() != "error validation" {
			fmt.Fprintf(w, "Error managementing: %v", err)
			return
		} else if err == nil {
			SendDocumentPage(w, r, data)
			return
		}
	}
	// путь откуда вызываю эту функцию, а не от её расположения
	tmplt, err := template.ParseFiles("../web/templates/index.html", "../web/templates/main.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing index.html and main.html: %v", err)
		return
	}
	// отображение страницы в заготовленном каркасе с footer
	err = tmplt.ExecuteTemplate(w, "index", wrongData)
	if err != nil {
		fmt.Fprintf(w, "Error displaying index.html and main.html: %v", err)
		return
	}
}

func ShowOptions(w http.ResponseWriter, r *http.Request) {
	var (
		options   []string
		parameter string = r.URL.Query().Get("parameter")
	)
	switch parameter {
	case "font":
		options = domain.Font
	case "fontsize":
		options = domain.Fontsize
	case "alignment":
		options = domain.Alignment
	case "spacing":
		options = domain.Spacing
	case "beforespacing":
		options = domain.BeforeSpacing
	case "afterspacing":
		options = domain.AfterSpacing
	case "firstindentation":
		options = domain.FirstIndentation
	case "listtabulation":
		options = domain.ListTabulation
	case "havetitle":
		options = domain.HaveTitle
	default:
		options = []string{}
	}

	// Возвращаем данные в формате JSON
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"options": options,
	})
	if err != nil {
		fmt.Fprintf(w, "Error sending json response: %v", err)
		return
	}
}

func SendDocumentPage(w http.ResponseWriter, r *http.Request, data domain.Answer) {
	var (
		fullData      domain.AnswerWithInterfaceName
		interfaceName string
	)

	// Загружаем и парсим шаблоны, добавляя пользовательскую функцию add
	tmplt, err := template.New("index.html").Funcs(template.FuncMap{
		"add": services.Add,
	}).ParseFiles("../web/templates/index.html", "../web/templates/download.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing download.html: %v", err)
		return
	}

	data.DocumentData.Filename = "formatted_" + data.DocumentData.Filename
	interfaceName = data.DocumentData.Filename[:10] + data.DocumentData.Filename[15:]
	// кодирование для URL
	data.DocumentData.Filename = url.QueryEscape(data.DocumentData.Filename)
	fullData = domain.AnswerWithInterfaceName{Data: data, InterfaceName: interfaceName}
	err = tmplt.ExecuteTemplate(w, "index", fullData)
	if err != nil {
		fmt.Fprintf(w, "Error displaying index.html and download.html: %v", err)
		return
	}
}

func SendDocument(w http.ResponseWriter, r *http.Request) {
	var (
		formattedDocumentName string
		formattedDocumentPath string
		interfaceName         string
		formattedDocument     *os.File
	)
	formattedDocumentName = r.URL.Query().Get("documentname")
	if formattedDocumentName == "" {
		fmt.Fprint(w, "Error getting the name of a new document from the page")
		return
	}
	// декодируем
	formattedDocumentName, err := url.QueryUnescape(formattedDocumentName)
	if err != nil {
		fmt.Fprintf(w, "Error decoding the name of a new document: %v", err)
		return
	}
	formattedDocumentPath = "../buffer/" + formattedDocumentName

	// проверка на наличие документа
	if _, err := os.Stat(formattedDocumentPath); err != nil {
		// перенаправление на страницу с ошибкой
		http.Redirect(w, r, "/error", http.StatusSeeOther)
		return
	}

	formattedDocument, err = os.Open(formattedDocumentPath)
	if err != nil {
		fmt.Fprintf(w, "Error opening the new document: %v", err)
		return
	}
	defer func() {
		formattedDocument.Close()
		infrastructure.DeleteBothDocuments(formattedDocumentName)
	}()

	// проверка на проблемы в файле
	formattedDocumentInfo, err := formattedDocument.Stat()
	if err != nil {
		fmt.Fprintf(w, "Error: problem in the document: %v", err)
		return
	}

	interfaceName = formattedDocumentName[:10] + formattedDocumentName[15:]

	// установка заголовков
	w.Header().Set("Content-Disposition", "attachment; filename="+interfaceName)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Length", strconv.Itoa(int(formattedDocumentInfo.Size())))

	// отдача файла пользователю
	_, err = io.Copy(w, formattedDocument)
	if err != nil {
		fmt.Fprintf(w, "Error sending new document: %v", err)
		return
	}
}

func ErrorPage(w http.ResponseWriter, r *http.Request) {
	tmplt, err := template.ParseFiles("../web/templates/index.html", "../web/templates/error.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing error.html: %v", err)
		return
	}
	err = tmplt.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Fprintf(w, "Error displaying index.html and error.html: %v", err)
		return
	}
}

func ErrorTimePage(w http.ResponseWriter, r *http.Request) {
	tmplt, err := template.ParseFiles("../web/templates/index.html", "../web/templates/errortime.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing errortime.html: %v", err)
		return
	}
	err = tmplt.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Fprintf(w, "Error displaying index.html and errortime.html: %v", err)
		return
	}
}
