package domain

import "mime/multipart"

// структура для передачи данных на сервере
type Answer struct {
	Document     multipart.File
	DocumentData *multipart.FileHeader
	Params       Parameters
	Information  DocumentInfo
	IsAllGood    []bool
}

type DocumentInfo struct {
	Draw           []bool
	Table          []bool
	Biblio         []bool
	ParagraphCount int
	Annotation     string
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
	HaveTitle        string `json:"haveTitle"`
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
	HaveTitle        []string = []string{"Есть", "Нет"}
)

// User profile structure
type User struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password,omitempty"`
}

// FormattingTemplate structure
type FormattingTemplate struct {
	ID               int64   `json:"id"`
	ProfileID        int64   `json:"profile_id"`
	Name             string  `json:"name"`
	Font             string  `json:"font"`
	Fontsize         int     `json:"fontsize"`
	Alignment        string  `json:"alignment"`
	Spacing          float64 `json:"spacing"`
	BeforeSpacing    float64 `json:"beforeSpacing"`
	AfterSpacing     float64 `json:"afterSpacing"`
	FirstIndentation float64 `json:"firstIndentation"`
	ListTabulation   float64 `json:"listTabulation"`
	HaveTitle        string  `json:"haveTitle"`
}

// API Request/Response structures
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	User    *User        `json:"user,omitempty"`
	Profile *UserProfile `json:"profile,omitempty"`
}

type TemplateResponse struct {
	Success   bool                 `json:"success"`
	Message   string               `json:"message"`
	Template  *FormattingTemplate  `json:"template,omitempty"`
	Templates []FormattingTemplate `json:"templates,omitempty"`
}

type UserProfile struct {
	UserID             int64                `json:"user_id"`
	Login              string               `json:"login"`
	Templates          []FormattingTemplate `json:"templates"`
	SelectedTemplateID int64                `json:"selected_template_id,omitempty"`
}
