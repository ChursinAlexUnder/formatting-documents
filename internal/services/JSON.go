package services

import (
	"encoding/json"
	"fmt"
	"formatting-documents/internal/domain"
	"io/ioutil"
	"sync"
	"time"
)

var mu sync.Mutex

func CheckDataJSON() error {
	var (
		filename string      = "data.json" // Имя файла
		data     domain.Data               // Переменная для данных
	)
	mu.Lock()         // Захватываем мьютекс
	defer mu.Unlock() // Освобождаем мьютекс после выхода из функции

	currentDate := time.Now().Format("2006-01-02") // Получаем текущую дату

	// Чтение файла через ioutil
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading json file: %v", err)
	}

	// Декодируем JSON в структуру
	err = json.Unmarshal(file, &data)
	if err != nil {
		return fmt.Errorf("error decoding json file: %v", err)
	}

	// Если дата устарела — обнуляем данные
	if data.Date != currentDate {
		data.Count = 0
		data.Date = currentDate
		data.LastFormatting = []domain.Parameters{}

		// Кодируем данные обратно в JSON
		updatedFile, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return fmt.Errorf("error coding json: %v", err)
		}

		// Записываем файл через ioutil
		err = ioutil.WriteFile(filename, updatedFile, 0644)
		if err != nil {
			return fmt.Errorf("error writing in json file: %v", err)
		}
		fmt.Println("JSON file updated successfully")
	}
	return nil
}
