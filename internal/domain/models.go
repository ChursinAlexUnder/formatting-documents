package domain

import "mime/multipart"

// структура для передачи данных на сервере
type Answer struct {
	Document     multipart.File
	DocumentData *multipart.FileHeader
	Params       Parameters
}

type Parameters struct {
	Font             string
	Fontsize         string
	Alignment        string
	Spacing          string
	Beforespacing    string
	Afterspacing     string
	Firstindentation string
}

// структура для отправки на страницу пользователю перез скачиванием
type AnswerWithInterfaceName struct {
	Data          Answer
	InterfaceName string
}

//  структура для валидации полей
type WrongData struct {
	ErrorDecorationButton     string
	ErrorCommentButton        string
	ErrorDecorationParameters string
	ErrorCommentParameters    string
}

// неизменяемые массивы!
var (
	Font             []string = []string{"Arial", "Times New Roman", "Calibri", "Courier New", "Verdana", "Georgia", "Tahoma"}
	Fontsize         []string = []string{"8", "9", "10", "11", "12", "13", "14", "16", "18", "20"}
	Alignment        []string = []string{"По левому краю", "По центру", "По правому краю", "По ширине"}
	Spacing          []string = []string{"1.0", "1.5", "2.0", "2.5", "3.0"}
	Beforespacing    []string = []string{"Нет", "1.0", "1.5", "2.0", "2.5", "3.0"}
	Afterspacing     []string = []string{"Нет", "1.0", "1.5", "2.0", "2.5", "3.0"}
	Firstindentation []string = []string{"0", "1.0", "1.25", "1.5", "1.75", "2.0", "2.5", "3.0"}
)
