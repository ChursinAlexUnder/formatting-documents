package infrastructure

import (
	"fmt"
	"formatting-documents/internal/domain"
	"io"
	"os"
)

// сохранение файла на сервере в папке buffer
func SaveDocument(data domain.Answer) error {
	var (
		bufferPath string = "./../buffer/" + data.DocumentData.Filename
	)
	downloadDocument, err := os.Create(bufferPath)
	if err != nil {
		return fmt.Errorf("error creating new empty document: %v", err)
	}
	defer downloadDocument.Close()

	_, err = io.Copy(downloadDocument, data.Document)
	if err != nil {
		return fmt.Errorf("error writing new emptry document: %v", err)
	}
	return nil
}
