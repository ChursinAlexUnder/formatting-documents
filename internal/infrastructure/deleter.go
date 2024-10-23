package infrastructure

import (
	"log"
	"os"
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
