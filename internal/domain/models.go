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
	BeforeSpacing    string
	AfterSpacing     string
	FirstIndentation string
	ListTabulation   string
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
	BeforeSpacing    []string = []string{"Нет", "1.0", "1.5", "2.0", "2.5", "3.0"}
	AfterSpacing     []string = []string{"Нет", "1.0", "1.5", "2.0", "2.5", "3.0"}
	FirstIndentation []string = []string{"0", "0.5", "1.0", "1.25", "1.5", "1.75", "2.0", "2.5", "3.0"}
	ListTabulation   []string = []string{"0", "0.25", "0.5", "0.75", "1.0", "1.25", "1.5", "1.75", "2.0", "2.25", "2.5", "2.75", "3.0", "3.25", "3.5", "3.75", "4.0"}
)
