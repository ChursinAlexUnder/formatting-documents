package infrastructure

import (
	"fmt"
	"formatting-documents/internal/config"
	"formatting-documents/internal/domain"
	"io"
	"os"
	"path/filepath"
)

func SaveDocument(data domain.Answer) error {
	var (
		bufferPath string = filepath.Join(config.BufferDir(), data.DocumentData.Filename)
	)
	downloadDocument, err := os.Create(bufferPath)
	if err != nil {
		return fmt.Errorf("не удалось создать файл документа: %v", err)
	}
	defer downloadDocument.Close()

	_, err = io.Copy(downloadDocument, data.Document)
	if err != nil {
		return fmt.Errorf("не удалось записать документ: %v", err)
	}
	return nil
}
