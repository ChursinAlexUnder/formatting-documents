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
	// путь откуда вызываю эту функцию, а не от её расположения
	tmplt, err := template.ParseFiles("../web/templates/index.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing index.html: %v", err)
		return
	}
	tmplt.Execute(w, nil)
}

func SendDocumentPage(w http.ResponseWriter, r *http.Request) {
	var (
		data          domain.Answer
		fullData      domain.AnswerWithInterfaceName
		interfaceName string
	)
	tmplt, err := template.ParseFiles("../web/templates/download.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing download.html: %v", err)
		return
	}
	data, err = ManagementData(w, r)
	if err != nil {
		fmt.Fprintf(w, "Error managementing: %v", err)
		return
	}
	data.DocumentData.Filename = "formatted_" + data.DocumentData.Filename
	interfaceName = data.DocumentData.Filename[:10] + data.DocumentData.Filename[15:]
	fullData = domain.AnswerWithInterfaceName{Data: data, InterfaceName: interfaceName}
	tmplt.Execute(w, fullData)
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
	tmplt, err := template.ParseFiles("../web/templates/error.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing error.html: %v", err)
		return
	}
	tmplt.Execute(w, nil)
}
