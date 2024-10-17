// package filework

// import (
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"log"
// 	"mime/multipart"
// 	"net/http"
// 	"os"
// 	"text/template"
// )

// type Answer struct {
// 	FileName, Change string
// }

// // для back4app
// func FormSend(w http.ResponseWriter, r *http.Request) {
// 	// Получаем файл и данные формы
// 	var change string
// 	file, fileData, err := r.FormFile("document-file")
// 	if err != nil {
// 		fmt.Fprintf(w, "Error not get file: %v", err)
// 		return
// 	}
// 	defer file.Close()
// 	change = r.FormValue("change")

// 	// Отправляем файл на Python микросервис для обработки
// 	editedFileName, err := sendFileToPythonService(file, fileData.Filename, change)
// 	if err != nil {
// 		fmt.Fprintf(w, "Error not send file: %v", err)
// 		return
// 	}

// 	// Загрузка и распарсивание HTML-шаблона
// 	tmp, err := template.ParseFiles("./html/temp.html")
// 	if err != nil {
// 		fmt.Fprintf(w, "Error not parse temp.html: %v", err)
// 		return
// 	}

// 	// Добавление данных на страницу
// 	answer := Answer{editedFileName, change}
// 	tmp.Execute(w, answer)
// }

// // для back4app
// func sendFileToPythonService(file multipart.File, filename string, comment string) (string, error) {
// 	// Создаем временный файл для сохранения загруженного контента
// 	tempFile, err := os.CreateTemp("", "upload-*.docx")
// 	if err != nil {
// 		return "", fmt.Errorf("error not create temp file: %v", err)
// 	}
// 	defer tempFile.Close()

// 	// Копируем содержимое загруженного файла в временный файл
// 	_, err = io.Copy(tempFile, file)
// 	if err != nil {
// 		return "", fmt.Errorf("error not copy file content: %v", err)
// 	}

// 	// Открываем временный файл для отправки
// 	tempFile, err = os.Open(tempFile.Name())
// 	if err != nil {
// 		return "", fmt.Errorf("error not open temp file: %v", err)
// 	}
// 	defer os.Remove(tempFile.Name()) // Удаляем временный файл после завершения

// 	// Создаем multipart-запрос
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)

// 	// Добавляем файл в запрос
// 	part, err := writer.CreateFormFile("file", filename)
// 	if err != nil {
// 		return "", fmt.Errorf("error not create form file: %v", err)
// 	}
// 	_, err = io.Copy(part, tempFile)
// 	if err != nil {
// 		return "", fmt.Errorf("error not copy file to form: %v", err)
// 	}

// 	// Добавляем комментарий в запрос
// 	err = writer.WriteField("comment", comment)
// 	if err != nil {
// 		return "", fmt.Errorf("error not add comment: %v", err)
// 	}

// 	writer.Close()
// 	// Отправляем файл на Python микросервис
// 	resp, err := http.Post("http://python:5000/editdocx", writer.FormDataContentType(), body)
// 	if err != nil {
// 		return "", fmt.Errorf("error not send request to Python service: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	// Сохраняем измененный файл
// 	editedFileName := "edited_" + filename
// 	outFile, err := os.Create(editedFileName)
// 	if err != nil {
// 		return "", fmt.Errorf("error not create output file: %v", err)
// 	}
// 	defer outFile.Close()

// 	_, err = io.Copy(outFile, resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("error not save edited file: %v", err)
// 	}

// 	return editedFileName, nil
// }

// func DownloadFile(w http.ResponseWriter, r *http.Request) {
// 	// Извлекаем имя файла из параметра URL
// 	fileName := r.URL.Query().Get("filename")
// 	if fileName == "" {
// 		http.Error(w, "File not found", http.StatusNotFound)
// 		return
// 	}

// 	// Открываем файл для чтения
// 	filePath := "./" + fileName
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		http.Error(w, "Error opening file", http.StatusInternalServerError)
// 		return
// 	}
// 	// Удаляем файл после успешного скачивания
// 	defer func() {
// 		err = os.Remove(filePath)
// 		if err != nil {
// 			log.Printf("Error deleting file %s: %v", fileName, err)
// 		} else {
// 			log.Printf("File %s successfully deleted", fileName)
// 		}
// 	}()
// 	// Закрываем файл
// 	defer file.Close()

// 	// Получаем информацию о файле
// 	fileInfo, err := file.Stat()
// 	if err != nil {
// 		http.Error(w, "Error getting file info", http.StatusInternalServerError)
// 		return
// 	}

// 	// Устанавливаем правильные заголовки для передачи файла
// 	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
// 	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
// 	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

// 	// Отправляем файл пользователю
// 	_, err = io.Copy(w, file)
// 	if err != nil {
// 		http.Error(w, "Error downloading file", http.StatusInternalServerError)
// 		return
// 	}
// }

package filework

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// Answer структура для передачи данных в шаблон
type Answer struct {
	FileName string
	Change   string
}

// FormSend обрабатывает загрузку файла и запуск Python-скрипта
func FormSend(w http.ResponseWriter, r *http.Request) {
	// Получаем файл и данные формы
	file, fileHeader, err := r.FormFile("document-file")
	if err != nil {
		fmt.Fprintf(w, "Error getting file: %v", err)
		return
	}
	defer file.Close()

	change := r.FormValue("change")

	// Определяем путь к папке python
	pythonDir, err := filepath.Abs("./python")
	if err != nil {
		fmt.Fprintf(w, "Error determining python directory: %v", err)
		return
	}

	// Фиксированная временная папка tempFolder
	tempDir := filepath.Join(pythonDir, "tempFolder")
	err = os.MkdirAll(tempDir, os.ModePerm) // Создаем папку tempFolder, если она не существует
	if err != nil {
		fmt.Fprintf(w, "Error creating temp directory: %v", err)
		return
	}

	// Сохраняем загруженный файл во временную папку
	originalFilePath := filepath.Join(tempDir, fileHeader.Filename)
	outFile, err := os.Create(originalFilePath)
	if err != nil {
		fmt.Fprintf(w, "Error creating file: %v", err)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		fmt.Fprintf(w, "Error saving file: %v", err)
		return
	}

	// Запускаем Python-скрипт
	editedFileName, err := runPythonScript(tempDir, pythonDir, fileHeader.Filename, change)
	if err != nil {
		fmt.Fprintf(w, "Error running Python script: %v", err)
		return
	}

	// Загрузка и распарсивание HTML-шаблона
	tmp, err := template.ParseFiles("./html/temp.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing temp.html: %v", err)
		return
	}

	// Добавление данных на страницу
	answer := Answer{editedFileName, change}
	tmp.Execute(w, answer)
}

// runPythonScript выполняет Python-скрипт для обработки файла
func runPythonScript(tempDir, pythonDir, filename, comment string) (string, error) {
	pythonScript := "editdocument.py" // Имя вашего Python-скрипта
	scriptPath := filepath.Join(pythonDir, pythonScript)

	// Проверяем, существует ли скрипт
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return "", fmt.Errorf("python script not found at %s", scriptPath)
	}

	// Полный путь к файлу, передаваемому в Python скрипт
	fullFilePath := filepath.Join(tempDir, filename)

	// Команда для выполнения скрипта
	cmd := exec.Command("python", scriptPath, fullFilePath, comment)

	// Устанавливаем рабочую директорию в pythonDir (это место, где исполняется скрипт)
	cmd.Dir = pythonDir

	// Получаем вывод и ошибки
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("python script error: %v, Output: %s", err, string(output))
	}

	// Предполагаем, что отредактированный файл называется "edited_<original>"
	editedFileName := "edited_" + filename
	editedFilePath := filepath.Join(tempDir, editedFileName)

	// Проверяем, что отредактированный файл существует
	if _, err := os.Stat(editedFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("edited file not found: %s", editedFilePath)
	}

	return editedFileName, nil
}

// DownloadFile предоставляет файл для скачивания
func DownloadFile(w http.ResponseWriter, r *http.Request) {
	// Получаем имя файла из запроса
	fileName := r.URL.Query().Get("filename")
	if fileName == "" {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Формируем полный путь к файлу
	filePath := "./python/tempFolder/" + fileName

	// Проверяем, существует ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Открываем файл для чтения
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

		// Удаляем оригинальный файл
		originalFileName := strings.Replace(fileName, "edited_", "", 1)
		originalFilePath := "./python/tempFolder/" + originalFileName
		err = os.Remove(originalFilePath)
		if err != nil {
			log.Printf("Error deleting original file %s: %v", originalFileName, err)
		} else {
			log.Printf("Original file %s successfully deleted", originalFileName)
		}

		// Проверяем, пуста ли папка после удаления файла
		dir := "./python/tempFolder"
		if isEmpty, err := isDirEmpty(dir); err == nil && isEmpty {
			err = os.Remove(dir)
			if err != nil {
				log.Printf("Error deleting folder %s: %v", dir, err)
			} else {
				log.Printf("Folder %s successfully deleted", dir)
			}
		}
	}()
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

// isDirEmpty проверяет, пуста ли директория
func isDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1) // Читаем первый файл в директории
	if err == io.EOF {
		return true, nil // Директория пуста
	}
	return false, err
}
