package infrastructure

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func DeleteDocument(documentName string) {
	var (
		documentPath string = "../buffer/" + documentName
		err          error
	)
	err = os.Remove(documentPath)
	if err != nil {
		log.Printf("Error deleting document %s: %v", documentName, err)
	} else {
		log.Printf("The document %s was successfully deleted", documentName)
	}
}

func DeleteOldDocuments() error {
	var (
		currentTime         time.Time = time.Now()
		bufferPath          string    = "../buffer"
		timeLastModDocument time.Time
		maxTimeStore        time.Duration = time.Minute * 10
	)
	documents, err := ioutil.ReadDir(bufferPath)
	if err != nil {
		return err
	}
	for _, document := range documents {
		timeLastModDocument = document.ModTime()
		if currentTime.Sub(timeLastModDocument) > maxTimeStore && document.Name() != ".gitkeep" {
			DeleteDocument(document.Name())
		}
	}
	return nil
}

// удаление документа и соответствующего отформатированного документа на сервере
func DeleteBothDocuments(formattedDocumentName string) {
	var documentName string
	DeleteDocument(formattedDocumentName)
	documentName = strings.Replace(formattedDocumentName, "formatted_", "", 1)
	DeleteDocument(documentName)
}
