package services

import (
	"encoding/json"
	"fmt"
	"formatting-documents/internal/config"
	"formatting-documents/internal/domain"
	"os"
	"sync"
	"time"
)

var mu sync.RWMutex

func ReadFileJSON(filename string) (domain.Data, error) {
	var data domain.Data

	mu.RLock()
	defer mu.RUnlock()

	file, err := os.ReadFile(filename)
	if err != nil {
		return data, fmt.Errorf("не удалось прочитать JSON-файл: %v", err)
	}

	err = json.Unmarshal(file, &data)
	if err != nil {
		return data, fmt.Errorf("не удалось декодировать JSON-файл: %v", err)
	}

	return data, nil
}

func WriteFileJSON(data domain.Data, filename string) error {
	updatedFile, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("не удалось сформировать JSON: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	err = os.WriteFile(filename, updatedFile, 0644)
	if err != nil {
		return fmt.Errorf("не удалось записать JSON-файл: %v", err)
	}

	return nil
}

func CheckDataJSON() error {
	var (
		filename string = config.DataFile()
		data     domain.Data
	)

	currentDate := time.Now().Add(3 * time.Hour).Format("2006-01-02")

	data, err := ReadFileJSON(filename)
	if err != nil {
		return fmt.Errorf("не удалось получить данные статистики: %v", err)
	}

	if data.Date != currentDate {
		data.Count = 0
		data.Date = currentDate
		data.LastFormatting = []domain.Parameters{}

		err = WriteFileJSON(data, filename)
		if err != nil {
			return fmt.Errorf("не удалось сбросить ежедневную статистику: %v", err)
		}
	}
	return nil
}

func UpdateDataJSON(params domain.Parameters) error {
	var (
		filename string = config.DataFile()
		data     domain.Data
	)

	data, err := ReadFileJSON(filename)
	if err != nil {
		return fmt.Errorf("не удалось получить статистику для обновления: %v", err)
	}

	data.Count++
	params.Time = time.Now().Add(3 * time.Hour).Format("15:04")

	data.LastFormatting = append([]domain.Parameters{params}, data.LastFormatting...)
	if len(data.LastFormatting) > 5 {
		data.LastFormatting = data.LastFormatting[:5]
	}

	err = WriteFileJSON(data, filename)
	if err != nil {
		return fmt.Errorf("не удалось сохранить статистику: %v", err)
	}

	return nil
}
