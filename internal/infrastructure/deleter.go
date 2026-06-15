package infrastructure

import (
	"formatting-documents/internal/config"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func DeleteDocument(documentName string) {
	var (
		documentPath string = filepath.Join(config.BufferDir(), documentName)
		err          error
	)
	err = os.Remove(documentPath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Не удалось удалить документ %s: %v", documentName, err)
	}
}

func DeleteOldDocuments() error {
	var (
		currentTime         time.Time = time.Now()
		bufferPath          string    = config.BufferDir()
		timeLastModDocument time.Time
		maxTimeStore        time.Duration = time.Minute * 10
	)
	documents, err := os.ReadDir(bufferPath)
	if err != nil {
		return err
	}
	for _, document := range documents {
		info, err := document.Info()
		if err != nil {
			return err
		}
		timeLastModDocument = info.ModTime()
		if currentTime.Sub(timeLastModDocument) > maxTimeStore && document.Name() != ".gitkeep" {
			DeleteDocument(document.Name())
		}
	}
	return nil
}
func DeleteBothDocuments(formattedDocumentName string) {
	var documentName string
	DeleteDocument(formattedDocumentName)
	documentName = strings.Replace(formattedDocumentName, "formatted_", "", 1)
	DeleteDocument(documentName)
}
