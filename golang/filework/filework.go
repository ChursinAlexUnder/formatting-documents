package filework

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"text/template"
)

type Answer struct {
	FileName, Change string
}

func FormSend(w http.ResponseWriter, r *http.Request) {
	// получение значения полей
	var change string
	file, fileData, err := r.FormFile("document-file")
	if err != nil {
		fmt.Fprintf(w, "Error not get file: %v", err)
		return
	}
	defer file.Close()
	change = r.FormValue("change")

	// Отправляем файл на Python микросервис для обработки
	editedFileName, err := sendFileToPythonService(file, fileData.Filename)
	if err != nil {
		fmt.Fprintf(w, "Error not send file: %v", err)
		return
	}

	// загрузка и распарсивание HTML-шаблона из файла temp.html для дальнейшей работы
	tmp, err := template.ParseFiles("./html/temp.html")
	if err != nil {
		fmt.Fprintf(w, "Error not parse temp.html: %v", err)
		return
	}

	// добавление данных на страницу
	answer := Answer{editedFileName, change}
	tmp.Execute(w, answer)
}

// Отправка файла на Python микросервис и получение измененного файла
func sendFileToPythonService(file multipart.File, filename string) (string, error) {
	// Создаем временный файл для сохранения загруженного контента
	tempFile, err := os.CreateTemp("", "upload-*.docx")
	if err != nil {
		return "", fmt.Errorf("error not create temp file: %v", err)
	}
	defer tempFile.Close()

	// Копируем содержимое загруженного файла в временный файл
	_, err = io.Copy(tempFile, file)
	if err != nil {
		return "", fmt.Errorf("error not copy file content: %v", err)
	}

	// Открываем временный файл для отправки
	tempFile, err = os.Open(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("error not open temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Удаляем временный файл после завершения

	// Создаем multipart-запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("error not create form file: %v", err)
	}

	_, err = io.Copy(part, tempFile)
	if err != nil {
		return "", fmt.Errorf("error not copy file to form: %v", err)
	}

	writer.Close()

	// Отправляем файл на Python микросервис
	resp, err := http.Post("http://python:5000/editdocx", writer.FormDataContentType(), body)
	if err != nil {
		return "", fmt.Errorf("error not send request to Python service: %v", err)
	}
	defer resp.Body.Close()

	// Сохраняем измененный файл
	editedFileName := "edited_" + filename
	outFile, err := os.Create(editedFileName)
	if err != nil {
		return "", fmt.Errorf("error not create output file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error not save edited file: %v", err)
	}

	return editedFileName, nil
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	// Извлекаем имя файла из параметра URL
	fileName := r.URL.Query().Get("filename")
	if fileName == "" {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Открываем файл для чтения
	filePath := "./" + fileName
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Error opening file", http.StatusInternalServerError)
		return
	}
	// Удаляем файл после успешного скачивания
	defer func() {
		err = os.Remove(filePath)
		if err != nil {
			log.Printf("Error deleting file %s: %v", fileName, err)
		} else {
			log.Printf("File %s successfully deleted", fileName)
		}
	}()
	// Закрываем файл
	defer file.Close()

	// Получаем информацию о файле
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Error getting file info", http.StatusInternalServerError)
		return
	}

	// Устанавливаем правильные заголовки для передачи файла
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Отправляем файл пользователю
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error downloading file", http.StatusInternalServerError)
		return
	}
}
