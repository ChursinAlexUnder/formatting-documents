package services

import "io/ioutil"

// считает размер папки buffer для предотвращения переполнения
func GetFolderSize() (int, error) {
	var (
		folderSize int64
		folderPath string = "../buffer"
	)

	// Читаем все файлы в папке
	documents, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return -1, err
	}

	for _, document := range documents {
		// Добавляем размер файла к общему размеру
		folderSize += document.Size()
	}

	return int(folderSize), nil
}
