package infrastructure

import (
	"io/ioutil"
	"log"
	"os"
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

func DeleteOldDocument() error {
	var (
		currentTime         time.Time = time.Now()
		bufferPath          string    = "../buffer"
		timeLastModDocument time.Time
		maxTimeStore        time.Duration = time.Minute * 1
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
