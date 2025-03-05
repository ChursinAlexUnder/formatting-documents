package services

import (
	"encoding/json"
	"fmt"
	"formatting-documents/internal/domain"
	"io/ioutil"
	"sync"
	"time"
)

var mu sync.RWMutex

func ReadFileJSON(filename string) (domain.Data, error) {
	var data domain.Data

	mu.RLock()
	defer mu.RUnlock()

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return data, fmt.Errorf("error reading JSON file: %v", err)
	}

	err = json.Unmarshal(file, &data)
	if err != nil {
		return data, fmt.Errorf("error decoding JSON file: %v", err)
	}

	return data, nil
}

func WriteFileJSON(data domain.Data, filename string) error {
	updatedFile, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("error coding JSON: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	err = ioutil.WriteFile(filename, updatedFile, 0644)
	if err != nil {
		return fmt.Errorf("error writing in json file: %v", err)
	}

	return nil
}

func CheckDataJSON() error {
	var (
		filename string = "../data.json"
		data     domain.Data
	)

	currentDate := time.Now().Format("2006-01-02")

	data, err := ReadFileJSON(filename)
	if err != nil {
		return fmt.Errorf("error getting information from JSON file for CHECK: %v", err)
	}

	if data.Date != currentDate {
		data.Count = 0
		data.Date = currentDate
		data.LastFormatting = []domain.Parameters{}

		err = WriteFileJSON(data, filename)
		if err != nil {
			return fmt.Errorf("error writing information in JSON file for CHECK: %v", err)
		}
	}
	return nil
}

func UpdateDataJSON(params domain.Parameters) error {
	var (
		filename string = "../data.json"
		data     domain.Data
	)

	data, err := ReadFileJSON(filename)
	if err != nil {
		return fmt.Errorf("error getting information from JSON file for UPDATE: %v", err)
	}

	data.Count++

	data.LastFormatting = append([]domain.Parameters{params}, data.LastFormatting...)
	if len(data.LastFormatting) > 5 {
		data.LastFormatting = data.LastFormatting[:5]
	}

	err = WriteFileJSON(data, filename)
	if err != nil {
		return fmt.Errorf("error writing information in JSON file for UPDATE: %v", err)
	}

	return nil
}
