package services

import (
	"fmt"
	"os"
	"os/exec"
)

func RunPythonScript(documentName, comment string) error {
	var (
		scriptPath            string = "../scripts/editdocument.py"
		formattedDocumentName string = "formatted_" + documentName
		directoryPath         string = "../scripts"
		bufferPath            string = "../buffer"
		cmd                   *exec.Cmd
	)

	// запуск скрипта
	cmd = exec.Command("python", scriptPath, documentName, comment)
	cmd.Dir = directoryPath

	// вывод ошибок от скрипта
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running python script: %v, output: %s", err, string(output))
	}

	if _, err := os.Stat(bufferPath + "/" + formattedDocumentName); err != nil {
		return fmt.Errorf("error creating formatted document: %v", err)
	}
	return nil
}
