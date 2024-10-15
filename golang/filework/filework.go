package filework

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"text/template"
	"time"
)

type Answer struct {
	FileName, Change string
}

var pythonCmd *exec.Cmd

func FormSend(w http.ResponseWriter, r *http.Request) {
	// Стартуем Python сервер перед отправкой файла
	err := startPythonServer()
	if err != nil {
		fmt.Fprintf(w, "Error not start python server: %v", err)
		return
	}

	// Получаем файл и данные формы
	var change string
	file, fileData, err := r.FormFile("document-file")
	if err != nil {
		fmt.Fprintf(w, "Error not get file: %v", err)
		return
	}
	defer file.Close()
	change = r.FormValue("change")

	// Отправляем файл на Python микросервис для обработки
	editedFileName, err := sendFileToPythonService(file, fileData.Filename, change)
	if err != nil {
		fmt.Fprintf(w, "Error not send file: %v", err)
		return
	}

	// Загрузка и распарсивание HTML-шаблона
	tmp, err := template.ParseFiles("./html/temp.html")
	if err != nil {
		fmt.Fprintf(w, "Error not parse temp.html: %v", err)
		return
	}

	// Добавление данных на страницу
	answer := Answer{editedFileName, change}
	tmp.Execute(w, answer)

	// Останавливаем Python сервер после получения файла
	stopPythonServer()
}

func sendFileToPythonService(file multipart.File, filename string, comment string) (string, error) {
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

	// Добавляем файл в запрос
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("error not create form file: %v", err)
	}
	_, err = io.Copy(part, tempFile)
	if err != nil {
		return "", fmt.Errorf("error not copy file to form: %v", err)
	}

	// Добавляем комментарий в запрос
	err = writer.WriteField("comment", comment)
	if err != nil {
		return "", fmt.Errorf("error not add comment: %v", err)
	}

	writer.Close()

	// Отправляем файл на Python микросервис
	resp, err := http.Post("http://localhost:5000/editdocx", writer.FormDataContentType(), body)
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

func startPythonServer() error {
	// Проверяем, что сервер ещё не запущен
	if pythonCmd != nil && pythonCmd.Process != nil {
		fmt.Println("Python сервер уже запущен")
		return nil
	}

	// Запуск Python-сервера
	pythonCmd = exec.Command("python", "python/editdocument.py")
	pythonCmd.Stdout = os.Stdout
	pythonCmd.Stderr = os.Stderr

	err := pythonCmd.Start()
	if err != nil {
		return fmt.Errorf("не удалось запустить Python сервер: %v", err)
	}

	// Ожидаем запуска сервера
	fmt.Println("Запуск Python сервера...")
	err = waitForPythonServer("http://localhost:5000")
	if err != nil {
		return fmt.Errorf("не удалось дождаться запуска Python сервера: %v", err)
	}

	fmt.Println("Python сервер запущен на порту 5000")
	return nil
}

func waitForPythonServer(url string) error {
	maxRetries := 10
	delay := 1000 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			// Если сервер ответил, завершаем ожидание
			return nil
		}

		// Ждем перед повторной проверкой
		time.Sleep(delay)
	}

	return fmt.Errorf("сервер не ответил после %d попыток", maxRetries)
}

func stopPythonServer() {
	// Проверяем, есть ли процесс для остановки
	if pythonCmd != nil && pythonCmd.Process != nil {
		err := pythonCmd.Process.Kill()
		if err != nil {
			fmt.Printf("Ошибка при остановке Python сервера: %v\n", err)
		} else {
			fmt.Println("Python сервер успешно остановлен")
		}
		pythonCmd = nil
	}
}
