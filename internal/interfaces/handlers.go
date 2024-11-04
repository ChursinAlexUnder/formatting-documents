package interfaces

import (
	"fmt"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/infrastructure"
	"io"
	"net/http"
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

func SendDocumentPage(w http.ResponseWriter, r *http.Request, data domain.Answer) {
	var (
		fullData      domain.AnswerWithInterfaceName
		interfaceName string
	)
	tmplt, err := template.ParseFiles("../web/templates/index.html", "../web/templates/download.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing download.html: %v", err)
		return
	}

	data.DocumentData.Filename = "formatted_" + data.DocumentData.Filename
	interfaceName = data.DocumentData.Filename[:10] + data.DocumentData.Filename[15:]
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
	formattedDocumentPath = "../buffer/" + formattedDocumentName

	// проверка на наличие документа
	if _, err := os.Stat(formattedDocumentPath); err != nil {
		// перенаправление на страницу с ошибкой
		http.Redirect(w, r, "/error", http.StatusSeeOther)
		return
	}

	formattedDocument, err := os.Open(formattedDocumentPath)
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
