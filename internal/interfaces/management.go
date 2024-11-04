package interfaces

import (
	"fmt"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/infrastructure"
	"formatting-documents/internal/services"
	"net/http"
	"regexp"
	"strings"
)

func ManagementData(w http.ResponseWriter, r *http.Request) (domain.Answer, domain.WrongData, error) {
	var (
		data      domain.Answer    = domain.Answer{}
		wrongData domain.WrongData = domain.WrongData{}
	)
	// удаление старых документов (которым больше 10 минут)
	err := infrastructure.DeleteOldDocuments()
	if err != nil {
		return data, wrongData, fmt.Errorf("error deleting old documents: %v", err)
	}
	// проверка на переполнение папки buffer
	err = services.IsOverflow()
	if err != nil {
		return data, wrongData, fmt.Errorf("error overflow: %v", err)
	}

	// валидация полей
	data, wrongData = Validation(r)
	if wrongData.ErrorDecorationButton != "" || wrongData.ErrorDecorationTextarea != "" {
		return data, wrongData, fmt.Errorf("error validation")
	}

	data = services.AddRandomNumber(data)

	// сохранение документа
	err = infrastructure.SaveDocument(data)
	if err != nil {
		return data, wrongData, fmt.Errorf("error saving the document on the server: %v", err)
	}

	// запуск python скрипта
	err = services.RunPythonScript(data.DocumentData.Filename, data.Comment)
	if err != nil {
		return data, wrongData, fmt.Errorf("error formatting the document on the server: %v", err)
	}
	return data, wrongData, nil
}

// проверка данных из формы
func Validation(r *http.Request) (domain.Answer, domain.WrongData) {
	const (
		maxDocumentSize int = 20 * 1024 * 1024
	)
	var (
		data      domain.Answer    = domain.Answer{}
		wrongData domain.WrongData = domain.WrongData{}
		comment   string
	)
	// получение данных из формы и валидация
	// документ
	document, documentHeader, err := r.FormFile("document-file")
	if err != nil || documentHeader.Filename == "" || document == nil {
		wrongData.ErrorDecorationButton = "-error"
		wrongData.ErrorCommentButton += "Документ обязательно необходимо загрузить."
	} else if !strings.HasSuffix(documentHeader.Filename, ".docx") {
		wrongData.ErrorDecorationButton = "-error"
		wrongData.ErrorCommentButton += "Для загрузки доступны документы только формата .docx."
	} else if int(documentHeader.Size) >= maxDocumentSize {
		wrongData.ErrorDecorationButton = "-error"
		wrongData.ErrorCommentButton += "Размер документа должен быть меньше 20 Мегабайт."
	} else if int(documentHeader.Size) == 0 {
		wrongData.ErrorDecorationButton = "-error"
		wrongData.ErrorCommentButton += "Документ не должен быть пустым."
	}
	if err == nil && document != nil {
		defer document.Close()
	}

	// комментарий
	comment = r.FormValue("change")
	flag, err := regexp.MatchString(`^\d+$`, comment)
	if err != nil {
		wrongData.ErrorDecorationTextarea = "-error"
		wrongData.ErrorCommentTextarea = "С комментарием что-то не так."
	} else if flag {
		wrongData.ErrorDecorationTextarea = "-error"
		wrongData.ErrorCommentTextarea = "Комментарий не может состоять только из цифр."
	}

	// если данные валидны, то сохраняем их в структуре
	if wrongData.ErrorDecorationButton == "" && wrongData.ErrorDecorationTextarea == "" {
		data = domain.Answer{Document: document, DocumentData: documentHeader, Comment: comment}
	}
	return data, wrongData
}
