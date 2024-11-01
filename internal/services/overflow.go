package services

import "io/ioutil"

// считает размер папки buffer для предотвращения переполнения
func GetBufferSize() (int, error) {
	var (
		bufferSize int64
		bufferPath string = "../buffer"
	)

	// Читаем все файлы в папке
	documents, err := ioutil.ReadDir(bufferPath)
	if err != nil {
		return -1, err
	}

	for _, document := range documents {
		// Добавляем размер файла к общему размеру
		bufferSize += document.Size()
	}

	return int(bufferSize), nil
}
