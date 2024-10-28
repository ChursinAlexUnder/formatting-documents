package services

import "os"

// считает размер папки buffer для предотвращения переполнения
func GetFolderSize() (int, error) {
	var (
		folderSize int64
		folderPath string = "../buffer"
	)

	// Читаем все файлы в папке
	documents, err := os.ReadDir(folderPath)
	if err != nil {
		return -1, err
	}

	for _, document := range documents {
		// Получаем информацию о файле и добавляем его размер
		information, err := document.Info()
		if err != nil {
			return -1, err
		}
		folderSize += information.Size() // Добавляем размер файла к общему размеру
	}

	return int(folderSize), nil
}
