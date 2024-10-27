package domain

import "mime/multipart"

// возможно, когда-нибудь сделать это всё через сессии

// отслеживание количества пользователей на сайте
var (
	Users []int = []int{}
	User  int   = 0
)

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
