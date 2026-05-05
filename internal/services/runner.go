package services

import (
	"encoding/json"
	"fmt"
	"formatting-documents/internal/domain"
	"os"
	"os/exec"
)

func RunPythonScript(documentName string, params domain.Parameters) (domain.DocumentInfo, error) {
	var (
		scriptPath            string = "../scripts/editdocument.py"
		formattedDocumentName string = "formatted_" + documentName
		directoryPath         string = "../scripts"
		bufferPath            string = "../buffer"
		cmd                   *exec.Cmd
	)

	// Получаем API ключ OpenRouter из переменной окружения
	openrouterApiKey := os.Getenv("OPENROUTER_API_KEY")

	// запуск скрипта
	cmd = exec.Command("python", scriptPath, documentName, params.Font, params.Fontsize, params.Alignment, params.Spacing, params.BeforeSpacing, params.AfterSpacing, params.FirstIndentation, params.ListTabulation, params.HaveTitle, openrouterApiKey)
	cmd.Dir = directoryPath

	// вывод ошибок от скрипта
	output, err := cmd.CombinedOutput()
	if err != nil {
		return domain.DocumentInfo{}, fmt.Errorf("error running python script: %v, output: %s", err, string(output))
	}

	if _, err := os.Stat(bufferPath + "/" + formattedDocumentName); err != nil {
		return domain.DocumentInfo{}, fmt.Errorf("error creating formatted document: %v", err)
	}

	// Парсим JSON-массив из 5 элементов
	var raw []interface{}
	err = json.Unmarshal(output, &raw)
	if err != nil {
		return domain.DocumentInfo{}, fmt.Errorf("error parsing json output: %v", err)
	}
	if len(raw) != 5 {
		return domain.DocumentInfo{}, fmt.Errorf("expected 5 elements in json array, got %d", len(raw))
	}

	drawList := convertToBoolSlice(raw[0])
	tableList := convertToBoolSlice(raw[1])
	biblioList := convertToBoolSlice(raw[2])
	paragraphCount := 0
	if val, ok := raw[3].(float64); ok {
		paragraphCount = int(val)
	}
	annotation := ""
	if val, ok := raw[4].(string); ok {
		annotation = val
	}

	return domain.DocumentInfo{
		Draw:           drawList,
		Table:          tableList,
		Biblio:         biblioList,
		ParagraphCount: paragraphCount,
		Annotation:     annotation,
	}, nil
}

func convertToBoolSlice(val interface{}) []bool {
	arr, ok := val.([]interface{})
	if !ok {
		return nil
	}
	result := make([]bool, len(arr))
	for i, v := range arr {
		if b, ok := v.(bool); ok {
			result[i] = b
		}
	}
	return result
}
