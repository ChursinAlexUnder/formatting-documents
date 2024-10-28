package domain

import "mime/multipart"

// возможно, когда-нибудь сделать это всё через сессии

// отслеживание пользователя на сайте
var (
	User int = 0
)

// структура для json файла
type Users struct {
	Numbers []int
}

// структура для передачи данных на сервере
type Answer struct {
	Document     multipart.File
	DocumentData *multipart.FileHeader
	Comment      string
}

type FullAnswer struct {
	Data     Answer
	TrueName string
}
