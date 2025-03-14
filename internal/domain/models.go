package domain

import "mime/multipart"

// структура для передачи данных на сервере
type Answer struct {
	Document     multipart.File
	DocumentData *multipart.FileHeader
	Params       Parameters
}

type Data struct {
	Count          int          `json:"count"`
	Date           string       `json:"date"`
	LastFormatting []Parameters `json:"last_formatting"`
}

type SSEData struct {
	Count          int          `json:"count"`
	LastFormatting []Parameters `json:"last_formatting"`
}

type Parameters struct {
	Time             string `json:"time"`
	Font             string `json:"font"`
	Fontsize         string `json:"fontsize"`
	Alignment        string `json:"alignment"`
	Spacing          string `json:"spacing"`
	BeforeSpacing    string `json:"beforeSpacing"`
	AfterSpacing     string `json:"afterSpacing"`
	FirstIndentation string `json:"firstIndentation"`
	ListTabulation   string `json:"listTabulation"`
	Content          string `json:"content"`
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
	BeforeSpacing    []string = []string{"0", "1.0", "1.5", "2.0", "2.5", "3.0"}
	AfterSpacing     []string = []string{"0", "1.0", "1.5", "2.0", "2.5", "3.0"}
	FirstIndentation []string = []string{"0", "0.5", "1.0", "1.25", "1.5", "1.75", "2.0", "2.5", "3.0"}
	ListTabulation   []string = []string{"0", "0.25", "0.5", "0.75", "1.0", "1.25", "1.5", "1.75", "2.0", "2.25", "2.5", "2.75", "3.0", "3.25", "3.5", "3.75", "4.0"}
	Content          []string = []string{"Добавить/обновить", "Не добавлять/не обновлять"}
)
