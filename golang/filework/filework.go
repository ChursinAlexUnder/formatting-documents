package filework

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type Answer struct {
	FileName, Text, Change string
}

func FormSend(w http.ResponseWriter, r *http.Request) {
	// получение значения полей
	var change string
	file, fileData, err := r.FormFile("document-file")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	defer file.Close()
	change = r.FormValue("change")

	// Проверка расширения файла на docx
	if ext := strings.ToLower(filepath.Ext(fileData.Filename)); ext != ".docx" {
		http.Error(w, "Invalid file extension", http.StatusUnsupportedMediaType)
		return
	}

	// загрузка и распарсивание HTML-шаблона из файла temp.html для дальнейшей работы
	tmp, err := template.ParseFiles("./html/temp.html")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	// Сохранение файла на диск
	tempFile, err := os.CreateTemp("", "temp-*.docx")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	// выполняться будут в обратном порядке
	defer os.Remove(tempFile.Name()) // Удаление файла после использования
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	// Вызов Python-скрипта для извлечения текста
	text, err := textFromDocx(tempFile.Name())
	if err != nil {
		fmt.Fprintf(w, "Error11111111: %v", err)
		return
	}

	// добавление данных на страницу
	answer := Answer{fileData.Filename, text, change}
	tmp.Execute(w, answer)
}

func textFromDocx(filename string) (string, error) {
	// Вызов Python-скрипта
	cmd := exec.Command("python3", "../../python/edit-document.py", filename)

	// Получение вывода из Python-скрипта
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Возврат текста как строки
	return strings.TrimSpace(string(output)), nil
}
