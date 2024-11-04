package domain

import "mime/multipart"

// структура для передачи данных на сервере
type Answer struct {
	Document     multipart.File
	DocumentData *multipart.FileHeader
	Comment      string
}

// структура для отправки на страницу пользователю перез скачиванием
type AnswerWithInterfaceName struct {
	Data          Answer
	InterfaceName string
}

//  структура для валидации полей
type WrongData struct {
	ErrorDecorationButton   string
	ErrorCommentButton      string
	ErrorDecorationTextarea string
	ErrorCommentTextarea    string
}
