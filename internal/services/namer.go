package services

import (
	"formatting-documents/internal/domain"
	"math/rand"
	"os"
	"strconv"
)

// добавление случайного кода к названию документа для уникальности
func AddRandomNumber(data domain.Answer) domain.Answer {
	var (
		randomNumber int
		path         string
		flag         bool = false
	)
	for !flag {
		randomNumber = rand.Intn(9000) + 1000
		data.DocumentData.Filename = strconv.Itoa(randomNumber) + "_" + data.DocumentData.Filename
		path = "../buffer/" + data.DocumentData.Filename
		if _, err := os.Stat(path); os.IsNotExist(err) {
			flag = true
		}
	}
	return data
}
