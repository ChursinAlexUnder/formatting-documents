package interfaces

import (
	"fmt"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/infrastructure"
	"formatting-documents/internal/services"
	"net/http"
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
		if err.Error() == "error: 6 iterations" {
			http.Redirect(w, r, "/errortime", http.StatusSeeOther)
		}
		return data, wrongData, fmt.Errorf("error overflow: %v", err)
	}

	// валидация полей
	data, wrongData = Validation(r)
	if wrongData.ErrorDecorationButton != "" || wrongData.ErrorDecorationParameters != "" {
		return data, wrongData, fmt.Errorf("error validation")
	}

	data = services.AddRandomNumber(data)

	// сохранение документа
	err = infrastructure.SaveDocument(data)
	if err != nil {
		return data, wrongData, fmt.Errorf("error saving the document on the server: %v", err)
	}

	// запуск python скрипта
	err = services.RunPythonScript(data.DocumentData.Filename, data.Params)
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
		data      domain.Answer     = domain.Answer{}
		wrongData domain.WrongData  = domain.WrongData{}
		params    domain.Parameters = domain.Parameters{}
	)
	// получение данных из формы и валидация
	// документ
	document, documentHeader, err := r.FormFile("document-file")
	if err != nil || documentHeader.Filename == "" || document == nil {
		wrongData.ErrorDecorationButton = "-error"
		wrongData.ErrorCommentButton += "Документ обязательно необходимо загрузить."
	} else if !strings.HasSuffix(documentHeader.Filename, ".docx") {
		wrongData.ErrorDecorationButton = "-error"
		wrongData.ErrorCommentButton += "Для загрузки доступны документы только формата docx."
	} else if len(documentHeader.Filename) > 50 {
		wrongData.ErrorDecorationButton = "-error"
		wrongData.ErrorCommentButton += "Название документа должно быть не длиннее 50 символов."
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

	// параметры
	params.Font = r.FormValue("font")
	params.Fontsize = r.FormValue("fontsize")
	params.Alignment = r.FormValue("alignment")
	params.Spacing = r.FormValue("spacing")
	params.Beforespacing = r.FormValue("beforespacing")
	params.Afterspacing = r.FormValue("afterspacing")
	params.Firstindentation = r.FormValue("firstindentation")

	if !services.InSlice(params.Font, domain.Font) {
		wrongData.ErrorDecorationParameters = "-error"
		wrongData.ErrorCommentParameters = "С шрифтом что-то не так."
	} else if !services.InSlice(params.Fontsize, domain.Fontsize) {
		wrongData.ErrorDecorationParameters = "-error"
		wrongData.ErrorCommentParameters = "С размером шрифта что-то не так."
	} else if !services.InSlice(params.Alignment, domain.Alignment) {
		wrongData.ErrorDecorationParameters = "-error"
		wrongData.ErrorCommentParameters = "С выравниванием текста что-то не так."
	} else if !services.InSlice(params.Spacing, domain.Spacing) {
		wrongData.ErrorDecorationParameters = "-error"
		wrongData.ErrorCommentParameters = "С междустрочным интервалом что-то не так."
	} else if !services.InSlice(params.Beforespacing, domain.Beforespacing) {
		wrongData.ErrorDecorationParameters = "-error"
		wrongData.ErrorCommentParameters = "С интервалом перед абзацем что-то не так."
	} else if !services.InSlice(params.Afterspacing, domain.Afterspacing) {
		wrongData.ErrorDecorationParameters = "-error"
		wrongData.ErrorCommentParameters = "С интервалом после абзаца что-то не так."
	} else if !services.InSlice(params.Firstindentation, domain.Firstindentation) {
		wrongData.ErrorDecorationParameters = "-error"
		wrongData.ErrorCommentParameters = "С отступом первой строки что-то не так."
	}

	// если данные валидны, то сохраняем их в структуре
	if wrongData.ErrorDecorationButton == "" && wrongData.ErrorDecorationParameters == "" {
		data = domain.Answer{Document: document, DocumentData: documentHeader, Params: params}
	}
	return data, wrongData
}
