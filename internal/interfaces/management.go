package interfaces

import (
	"fmt"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/infrastructure"
	"formatting-documents/internal/services"
	"net/http"
	"strconv"
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

	err = services.AddUserNumber()
	if err != nil {
		return data, fmt.Errorf("error adding a new user: %v", err)
	}

	data = domain.Answer{Document: document, DocumentData: documentHeader, Comment: comment}

	// добавление метки пользователя к названию документа
	data.DocumentData.Filename = strconv.Itoa(domain.User) + "_" + data.DocumentData.Filename

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
