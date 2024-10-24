package interfaces

import (
	"fmt"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/infrastructure"
	"formatting-documents/internal/services"
	"net/http"
)

func ManagementData(w http.ResponseWriter, r *http.Request) (domain.Answer, error) {
	// получение данных из формы
	var (
		comment string
		data    domain.Answer = domain.Answer{Document: nil, DocumentData: nil, Comment: ""}
	)
	document, documentHeader, err := r.FormFile("document-file")
	if err != nil {
		return data, fmt.Errorf("error getting document: %v", err)
	}
	defer document.Close()
	comment = r.FormValue("change")

	// сохранение документа
	data = domain.Answer{Document: document, DocumentData: documentHeader, Comment: comment}
	err = infrastructure.SaveDocument(data)
	if err != nil {
		return data, fmt.Errorf("error saving the document on the server: %v", err)
	}

	// запуск python скрипта
	err = services.RunPythonScript(data.DocumentData.Filename, data.Comment)
	if err != nil {
		return data, fmt.Errorf("error formatting the document on the server: %v", err)
	}

	data.Comment = comment
	data.Document = document
	data.DocumentData = documentHeader
	return data, nil
}
