package interfaces

import (
	"fmt"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/infrastructure"
	"formatting-documents/internal/services"
	"net/http"
)

func ManagementData(w http.ResponseWriter, r *http.Request) (domain.Answer, error) {
	var (
		comment string
		data    domain.Answer = domain.Answer{Document: nil, DocumentData: nil, Comment: ""}
	)
	// удаление старых документов (которым больше 10 минут)
	err := infrastructure.DeleteOldDocuments()
	if err != nil {
		return data, fmt.Errorf("error deleting old documents: %v", err)
	}
	// проверка на переполнение папки buffer
	err = services.IsOverflow()
	if err != nil {
		return data, fmt.Errorf("error overflow: %v", err)
	}
	// получение данных из формы
	document, documentHeader, err := r.FormFile("document-file")
	if err != nil {
		return data, fmt.Errorf("error getting document: %v", err)
	}
	defer document.Close()
	comment = r.FormValue("change")

	data = domain.Answer{Document: document, DocumentData: documentHeader, Comment: comment}

	data = services.AddRandomNumber(data)

	// сохранение документа
	err = infrastructure.SaveDocument(data)
	if err != nil {
		return data, fmt.Errorf("error saving the document on the server: %v", err)
	}

	// запуск python скрипта
	err = services.RunPythonScript(data.DocumentData.Filename, data.Comment)
	if err != nil {
		return data, fmt.Errorf("error formatting the document on the server: %v", err)
	}
	return data, nil
}
