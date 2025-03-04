package services

import (
	"encoding/json"
	"fmt"
	"formatting-documents/internal/domain"
	"os"
	"sync"
	"time"
)

var mu sync.Mutex

func CheckDataJSON() error {
	var (
		filename string = "data.json"
		data     domain.Data
	)
	mu.Lock()
	defer mu.Unlock()

	currentDate := time.Now().Format("2006-01-02")

	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading json file: %v", err)
	}

	err = json.Unmarshal(file, &data)
	if err != nil {
		return fmt.Errorf("error decoding json file: %v", err)
	}

	if data.Date != currentDate {
		data.Count = 0
		data.Date = currentDate
		data.LastFormatting = []domain.Parameters{}

		updatedFile, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return fmt.Errorf("error coding json: %v", err)
		}

		err = os.WriteFile(filename, updatedFile, 0644)
		if err != nil {
			return fmt.Errorf("error writing in json file: %v", err)
		}
	}
	return nil
}
