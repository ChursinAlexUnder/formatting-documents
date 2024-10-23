package interfaces

import (
	"fmt"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/infrastructure"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
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
		data domain.Answer
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
	tmplt.Execute(w, data)
}

func SendDocument(w http.ResponseWriter, r *http.Request) {
	var (
		formattedDocumentName string
		documentName          string
		formattedDocumentPath string
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
		fmt.Fprintf(w, "Error: problem with the new document: %v", err)
		return
	}

	formattedDocument, err := os.Open(formattedDocumentPath)
	if err != nil {
		fmt.Fprintf(w, "Error opening the new document: %v", err)
		return
	}
	defer func() {
		formattedDocument.Close()
		infrastructure.DeleteDocument(formattedDocumentName)
		documentName = strings.Replace(formattedDocumentName, "formatted_", "", 1)
		infrastructure.DeleteDocument(documentName)
	}()

	// проверка на проблемы в файле
	formattedDocumentInfo, err := formattedDocument.Stat()
	if err != nil {
		fmt.Fprintf(w, "Error: problem in the document: %v", err)
		return
	}

	// установка заголовков
	w.Header().Set("Content-Disposition", "attachment; filename="+formattedDocumentName)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Length", strconv.Itoa(int(formattedDocumentInfo.Size())))

	// отдача файла пользователю
	_, err = io.Copy(w, formattedDocument)
	if err != nil {
		fmt.Fprintf(w, "Error sending new document: %v", err)
		return
	}
}
