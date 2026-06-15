package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"formatting-documents/internal/config"
	"formatting-documents/internal/domain"
	"os"
	"os/exec"
	"path/filepath"
)

func RunPythonScript(documentName string, params domain.Parameters) (domain.DocumentInfo, error) {
	var (
		scriptPath            string = config.RootPath("scripts", "editdocument.py")
		formattedDocumentName string = "formatted_" + documentName
		directoryPath         string = config.RootPath("scripts")
		bufferPath            string = config.BufferDir()
		cmd                   *exec.Cmd
		stdout                bytes.Buffer
		stderr                bytes.Buffer
	)
	openrouterApiKey := os.Getenv("OPENROUTER_API_KEY")
	cmd = exec.Command(config.PythonBin(), scriptPath, documentName, params.Font, params.Fontsize, params.Alignment, params.Spacing, params.BeforeSpacing, params.AfterSpacing, params.FirstIndentation, params.ListTabulation, params.HaveTitle, openrouterApiKey)
	cmd.Dir = directoryPath
	cmd.Env = append(os.Environ(), "APP_BUFFER_DIR="+bufferPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return domain.DocumentInfo{}, fmt.Errorf(
			"ошибка запуска Python-скрипта: %v; stderr: %s; stdout: %s",
			err,
			stderr.String(),
			stdout.String(),
		)
	}
	output := stdout.Bytes()

	if _, err := os.Stat(filepath.Join(bufferPath, formattedDocumentName)); err != nil {
		return domain.DocumentInfo{}, fmt.Errorf("отформатированный документ не создан: %v", err)
	}

	var raw []interface{}
	err = json.Unmarshal(output, &raw)
	if err != nil {
		return domain.DocumentInfo{}, fmt.Errorf("не удалось разобрать JSON-ответ Python: %v", err)
	}
	if len(raw) != 5 {
		return domain.DocumentInfo{}, fmt.Errorf("ожидалось 5 элементов JSON, получено %d", len(raw))
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
