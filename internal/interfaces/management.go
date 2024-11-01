package interfaces

import (
	"fmt"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/infrastructure"
	"formatting-documents/internal/services"
	"net/http"
	"time"
)

func ManagementData(w http.ResponseWriter, r *http.Request) (domain.Answer, error) {
	const (
		maxFolderSize int = 200 * 1024 * 1024
	)
	var (
		comment    string
		folderSize int
		data       domain.Answer = domain.Answer{Document: nil, DocumentData: nil, Comment: ""}
	)

	// удаление старых документов (которым больше 10 минут)
	err := infrastructure.DeleteOldDocument()
	if err != nil {
		return data, fmt.Errorf("error deleting old documents: %v", err)
	}

	// проверка на переполнение
	folderSize, err = services.GetBufferSize()
	if err != nil {
		return data, fmt.Errorf("error getting folder buffer size: %v", err)
	}
	for folderSize >= maxFolderSize {
		time.Sleep(3 * time.Second)
		folderSize, err = services.GetBufferSize()
		if err != nil {
			return data, fmt.Errorf("error getting folder buffer size: %v", err)
		}
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
