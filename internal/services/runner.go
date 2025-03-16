package services

import (
	"encoding/json"
	"fmt"
	"formatting-documents/internal/domain"
	"os"
	"os/exec"
)

func RunPythonScript(documentName string, params domain.Parameters) error {
	var (
		scriptPath            string = "../scripts/editdocument.py"
		formattedDocumentName string = "formatted_" + documentName
		directoryPath         string = "../scripts"
		bufferPath            string = "../buffer"
		cmd                   *exec.Cmd
	)

	// запуск скрипта
	cmd = exec.Command("python", scriptPath, documentName, params.Font, params.Fontsize, params.Alignment, params.Spacing, params.BeforeSpacing, params.AfterSpacing, params.FirstIndentation, params.ListTabulation)
	cmd.Dir = directoryPath

	// вывод ошибок от скрипта
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running python script: %v, output: %s", err, string(output))
	}

	if _, err := os.Stat(bufferPath + "/" + formattedDocumentName); err != nil {
		return fmt.Errorf("error creating formatted document: %v", err)
	}

	// TODO: НАЛАДИТЬ ОБРАБОТКУ И ВЫВОД ДАННЫХ ПОЛЬЗОВАТЕЛЮ НА СТРАНИЦУ!!!!

	// Например, ожидаем, что Python вернул список массивов: [[...], [...], [...]]
	var result [][]bool
	err = json.Unmarshal(output, &result)
	if err != nil {
		return fmt.Errorf("error parsing json output from python script: %v", err)
	}

	// Теперь result содержит три массива, и вы можете их использовать
	fmt.Println("Parsed arrays:", result)

	return nil
}
