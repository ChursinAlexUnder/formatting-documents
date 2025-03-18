package services

import (
	"encoding/json"
	"fmt"
	"formatting-documents/internal/domain"
	"os"
	"os/exec"
)

func RunPythonScript(documentName string, params domain.Parameters) ([][]bool, error) {
	var (
		scriptPath            string = "../scripts/editdocument.py"
		formattedDocumentName string = "formatted_" + documentName
		directoryPath         string = "../scripts"
		bufferPath            string = "../buffer"
		cmd                   *exec.Cmd
		result                [][]bool
	)

	// запуск скрипта
	cmd = exec.Command("python", scriptPath, documentName, params.Font, params.Fontsize, params.Alignment, params.Spacing, params.BeforeSpacing, params.AfterSpacing, params.FirstIndentation, params.ListTabulation, params.HaveTitle)
	cmd.Dir = directoryPath

	// вывод ошибок от скрипта
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running python script: %v, output: %s", err, string(output))
	}

	if _, err := os.Stat(bufferPath + "/" + formattedDocumentName); err != nil {
		return nil, fmt.Errorf("error creating formatted document: %v", err)
	}

	// получение данных от python скрипта
	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing json output from python script: %v", err)
	}

	return result, nil
}
