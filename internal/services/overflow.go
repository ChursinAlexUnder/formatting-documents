package services

import (
	"formatting-documents/internal/infrastructure"
	"io/ioutil"
	"time"
)

// считает размер папки buffer для предотвращения переполнения
func GetBufferSize() (int, error) {
	var (
		bufferSize int64
		bufferPath string = "../buffer"
	)

	// Читаем все файлы в папке
	documents, err := ioutil.ReadDir(bufferPath)
	if err != nil {
		return -1, err
	}

	for _, document := range documents {
		// Добавляем размер файла к общему размеру
		bufferSize += document.Size()
	}

	return int(bufferSize), nil
}

// вычисление размера папки buffer и её ограничение на 200 Мегабайт
func IsOverflow() error {
	const (
		maxFolderSize int = 200 * 1024 * 1024
	)
	var (
		folderSize int
		err        error
	)
	folderSize, err = GetBufferSize()
	if err != nil {
		return err
	}
	for folderSize >= maxFolderSize {
		time.Sleep(3 * time.Second)
		// удаление старых документов (которым больше 10 минут)
		err := infrastructure.DeleteOldDocuments()
		if err != nil {
			return err
		}
		folderSize, err = GetBufferSize()
		if err != nil {
			return err
		}
	}
	return nil
}
