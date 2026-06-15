package services

import (
	"errors"
	"fmt"
	"formatting-documents/internal/config"
	"formatting-documents/internal/infrastructure"
	"os"
	"time"
)

var ErrBufferBusy = errors.New("хранилище временных документов занято")

func GetBufferSize() (int, error) {
	var (
		bufferSize int64
		bufferPath string = config.BufferDir()
	)

	documents, err := os.ReadDir(bufferPath)
	if err != nil {
		return -1, err
	}

	for _, document := range documents {
		info, err := document.Info()
		if err != nil {
			return -1, err
		}
		bufferSize += info.Size()
	}

	return int(bufferSize), nil
}

func IsOverflow() error {
	const (
		maxBufferSize int = 200 * 1024 * 1024
	)
	var (
		iterations int = 0
		bufferSize int
		err        error
	)
	bufferSize, err = GetBufferSize()
	if err != nil {
		return err
	}
	for bufferSize >= maxBufferSize {
		time.Sleep(3 * time.Second)
		err := infrastructure.DeleteOldDocuments()
		if err != nil {
			return err
		}
		bufferSize, err = GetBufferSize()
		if err != nil {
			return err
		}
		iterations++
		if iterations >= 6 {
			return fmt.Errorf("%w после шести попыток очистки", ErrBufferBusy)
		}
	}
	return nil
}
