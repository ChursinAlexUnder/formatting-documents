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
		filename string = "data.json"
		data     domain.Data
	)
	mu.Lock()
	defer mu.Unlock()

	currentDate := time.Now().Format("2006-01-02")

	file, err := ioutil.ReadFile(filename)
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

		err = ioutil.WriteFile(filename, updatedFile, 0644)
		if err != nil {
			return fmt.Errorf("error writing in json file: %v", err)
		}
	}
	return nil
}

func UpdateDataJSON(params domain.Parameters) error {
	var (
		filename string = "data.json"
		data     domain.Data
	)
	mu.Lock()
	defer mu.Unlock()

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading json file: %v", err)
	}

	err = json.Unmarshal(file, &data)
	if err != nil {
		return fmt.Errorf("error decoding json file: %v", err)
	}

	data.Count++

	data.LastFormatting = append([]domain.Parameters{params}, data.LastFormatting...)
	if len(data.LastFormatting) > 5 {
		data.LastFormatting = data.LastFormatting[:5]
	}

	updatedFile, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("error encoding json: %v", err)
	}

	err = ioutil.WriteFile(filename, updatedFile, 0644)
	if err != nil {
		return fmt.Errorf("error writing to json file: %v", err)
	}

	return nil
}
