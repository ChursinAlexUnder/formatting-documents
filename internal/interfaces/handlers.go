package interfaces

import (
	"encoding/json"
	"fmt"
	"formatting-documents/database"
	"formatting-documents/internal/config"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/infrastructure"
	"formatting-documents/internal/services"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"
)

const maxTemplateNameLength = 60

func getUserIDFromCookie(r *http.Request) (int64, error) {
	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		return 0, fmt.Errorf("пользователь не авторизован")
	}

	var userID int64
	if _, err := fmt.Sscanf(cookie.Value, "%d", &userID); err != nil || userID <= 0 {
		return 0, fmt.Errorf("недействительная сессия")
	}

	return userID, nil
}

func databaseUnavailableJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": "База данных недоступна",
	})
}

func TurnstileConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	siteKey, err := TurnstileSiteKey()
	if err != nil {
		log.Printf("Ошибка настройки Turnstile: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Капча не настроена",
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"site_key": siteKey,
	})
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		if cookie, err := r.Cookie("profile_welcome"); err == nil && cookie.Value == "1" {
			clearProfileWelcomeCookie(w)
		}
	}

	var (
		wrongData domain.WrongData = domain.WrongData{}
		data      domain.Answer
		err       error
	)
	err = services.CheckDataJSON()
	if err != nil {
		fmt.Fprintf(w, "Ошибка проверки данных JSON: %v", err)
		return
	}
	if r.Method == http.MethodPost {
		data, wrongData, err = ManagementData(w, r)
		if err != nil && err.Error() != "ошибка валидации" {
			fmt.Fprintf(w, "Ошибка обработки документа: %v", err)
			return
		} else if err == nil {
			SendDocumentPage(w, r, data)
			return
		}
	}
	tmplt, err := template.ParseFiles(
		config.RootPath("web", "templates", "index.html"),
		config.RootPath("web", "templates", "main.html"),
	)
	if err != nil {
		fmt.Fprintf(w, "Ошибка загрузки шаблонов главной страницы: %v", err)
		return
	}
	err = tmplt.ExecuteTemplate(w, "index", wrongData)
	if err != nil {
		fmt.Fprintf(w, "Ошибка отображения главной страницы: %v", err)
		return
	}
}

func ShowOptions(w http.ResponseWriter, r *http.Request) {
	var (
		options   []string
		parameter string = r.URL.Query().Get("parameter")
	)
	switch parameter {
	case "font":
		options = domain.Font
	case "fontsize":
		options = domain.Fontsize
	case "alignment":
		options = domain.Alignment
	case "spacing":
		options = domain.Spacing
	case "beforespacing":
		options = domain.BeforeSpacing
	case "afterspacing":
		options = domain.AfterSpacing
	case "firstindentation":
		options = domain.FirstIndentation
	case "listtabulation":
		options = domain.ListTabulation
	case "havetitle":
		options = domain.HaveTitle
	default:
		options = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"options": options,
	})
	if err != nil {
		fmt.Fprintf(w, "Ошибка отправки JSON-ответа: %v", err)
		return
	}
}

func SendDocumentPage(w http.ResponseWriter, r *http.Request, data domain.Answer) {
	var (
		fullData      domain.AnswerWithInterfaceName
		interfaceName string
	)
	tmplt, err := template.New("index.html").Funcs(template.FuncMap{
		"add": services.Add,
	}).ParseFiles(
		config.RootPath("web", "templates", "index.html"),
		config.RootPath("web", "templates", "download.html"),
	)
	if err != nil {
		fmt.Fprintf(w, "Ошибка загрузки страницы скачивания: %v", err)
		return
	}

	data.DocumentData.Filename = "formatted_" + data.DocumentData.Filename
	interfaceName = data.DocumentData.Filename[:10] + data.DocumentData.Filename[15:]
	data.DocumentData.Filename = url.QueryEscape(data.DocumentData.Filename)
	fullData = domain.AnswerWithInterfaceName{Data: data, InterfaceName: interfaceName}
	err = tmplt.ExecuteTemplate(w, "index", fullData)
	if err != nil {
		fmt.Fprintf(w, "Ошибка отображения страницы скачивания: %v", err)
		return
	}
}

func SendDocument(w http.ResponseWriter, r *http.Request) {
	var (
		formattedDocumentName string
		formattedDocumentPath string
		interfaceName         string
		formattedDocument     *os.File
	)
	formattedDocumentName = r.URL.Query().Get("documentname")
	if formattedDocumentName == "" {
		fmt.Fprint(w, "Не удалось получить имя документа")
		return
	}
	formattedDocumentName, err := url.QueryUnescape(formattedDocumentName)
	if err != nil {
		fmt.Fprintf(w, "Не удалось декодировать имя документа: %v", err)
		return
	}
	formattedDocumentName = filepath.Base(strings.ReplaceAll(formattedDocumentName, "\\", "/"))
	if !strings.HasPrefix(formattedDocumentName, "formatted_") {
		http.Redirect(w, r, "/error", http.StatusSeeOther)
		return
	}
	formattedDocumentPath = filepath.Join(config.BufferDir(), formattedDocumentName)
	if _, err := os.Stat(formattedDocumentPath); err != nil {
		http.Redirect(w, r, "/error", http.StatusSeeOther)
		return
	}

	formattedDocument, err = os.Open(formattedDocumentPath)
	if err != nil {
		fmt.Fprintf(w, "Не удалось открыть документ: %v", err)
		return
	}
	defer func() {
		formattedDocument.Close()
		infrastructure.DeleteBothDocuments(formattedDocumentName)
	}()
	formattedDocumentInfo, err := formattedDocument.Stat()
	if err != nil {
		fmt.Fprintf(w, "Ошибка чтения документа: %v", err)
		return
	}

	interfaceName = formattedDocumentName[:10] + formattedDocumentName[15:]
	w.Header().Set("Content-Disposition", "attachment; filename="+interfaceName)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Length", strconv.Itoa(int(formattedDocumentInfo.Size())))
	_, err = io.Copy(w, formattedDocument)
	if err != nil {
		fmt.Fprintf(w, "Ошибка отправки документа: %v", err)
		return
	}
}

func ErrorPage(w http.ResponseWriter, r *http.Request) {
	tmplt, err := template.ParseFiles(
		config.RootPath("web", "templates", "index.html"),
		config.RootPath("web", "templates", "error.html"),
		config.RootPath("web", "templates", "auth-menu.html"),
	)
	if err != nil {
		fmt.Fprintf(w, "Ошибка загрузки страницы ошибки: %v", err)
		return
	}
	err = tmplt.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Fprintf(w, "Ошибка отображения страницы ошибки: %v", err)
		return
	}
}

func ErrorTimePage(w http.ResponseWriter, r *http.Request) {
	tmplt, err := template.ParseFiles(
		config.RootPath("web", "templates", "index.html"),
		config.RootPath("web", "templates", "errortime.html"),
		config.RootPath("web", "templates", "auth-menu.html"),
	)
	if err != nil {
		fmt.Fprintf(w, "Ошибка загрузки страницы ожидания: %v", err)
		return
	}
	err = tmplt.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Fprintf(w, "Ошибка отображения страницы ожидания: %v", err)
		return
	}
}

func InfoPage(w http.ResponseWriter, r *http.Request) {
	tmplt, err := template.ParseFiles(
		config.RootPath("web", "templates", "index.html"),
		config.RootPath("web", "templates", "info.html"),
		config.RootPath("web", "templates", "auth-menu.html"),
	)
	if err != nil {
		fmt.Fprintf(w, "Ошибка загрузки страницы инструкции: %v", err)
		return
	}
	err = tmplt.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Fprintf(w, "Ошибка отображения страницы инструкции: %v", err)
		return
	}
}
func ProfilePage(w http.ResponseWriter, r *http.Request) {
	if !database.IsAvailable() {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	_, err := getUserIDFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	isFirstVisit := false
	if cookie, cookieErr := r.Cookie("profile_welcome"); cookieErr == nil && cookie.Value == "1" {
		isFirstVisit = true
	}

	tmplt, err := template.ParseFiles(
		config.RootPath("web", "templates", "index.html"),
		config.RootPath("web", "templates", "profile.html"),
	)
	if err != nil {
		fmt.Fprintf(w, "Ошибка загрузки страницы профиля: %v", err)
		return
	}
	err = tmplt.ExecuteTemplate(w, "index", map[string]bool{
		"IsFirstVisit": isFirstVisit,
	})
	if err != nil {
		fmt.Fprintf(w, "Ошибка отображения страницы профиля: %v", err)
		return
	}
}
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if !database.IsAvailable() {
		databaseUnavailableJSON(w)
		return
	}

	var req domain.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Некорректный запрос",
		})
		return
	}

	if len(req.Login) == 0 || len(req.Password) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
		})
		return
	}

	ip := ClientIP(r)
	validation, err := Validate(r.Context(), req.TurnstileToken, ip, "register")
	if err != nil || validation == nil || !validation.Success {
		log.Printf("Ошибка проверки капчи при регистрации: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Проверка капчи не пройдена",
		})
		return
	}

	user, err := database.CreateUser(req.Login, req.Password)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    fmt.Sprintf("%d", user.ID),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "profile_welcome",
		Value:    "1",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Регистрация выполнена успешно",
		"user_id": user.ID,
	})
}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if !database.IsAvailable() {
		databaseUnavailableJSON(w)
		return
	}

	var req domain.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Некорректный запрос",
		})
		return
	}

	ip := ClientIP(r)
	validation, err := Validate(r.Context(), req.TurnstileToken, ip, "login")
	if err != nil || validation == nil || !validation.Success {
		log.Printf("Ошибка проверки капчи при входе: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Проверка капчи не пройдена",
		})
		return
	}

	userID, err := database.VerifyPassword(req.Login, req.Password)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    fmt.Sprintf("%d", userID),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})
	clearProfileWelcomeCookie(w)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Вход выполнен успешно",
		"user_id": userID,
	})
}
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "selected_template",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteLaxMode,
	})
	clearProfileWelcomeCookie(w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

func clearProfileWelcomeCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "profile_welcome",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if !database.IsAvailable() {
		databaseUnavailableJSON(w)
		return
	}

	userID, err := getUserIDFromCookie(r)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Not authenticated",
		})
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Пользователь не найден",
		})
		return
	}

	templates, err := database.GetTemplatesByUserID(userID)
	if err != nil {
		templates = []domain.FormattingTemplate{}
	}
	selectedCookie, _ := r.Cookie("selected_template")
	selectedID := int64(0)
	if selectedCookie != nil {
		fmt.Sscanf(selectedCookie.Value, "%d", &selectedID)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":              true,
		"user_id":              user.ID,
		"login":                user.Login,
		"templates":            templates,
		"selected_template_id": selectedID,
	})
}
func CreateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if !database.IsAvailable() {
		databaseUnavailableJSON(w)
		return
	}

	userID, err := getUserIDFromCookie(r)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Not authenticated",
		})
		return
	}

	var template domain.FormattingTemplate
	err = json.NewDecoder(r.Body).Decode(&template)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Некорректный запрос",
		})
		return
	}

	if err = validateTemplateName(&template); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	template.ProfileID = userID

	createdTemplate, err := database.CreateTemplate(&template)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"template": createdTemplate,
	})
}
func GetTemplateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !database.IsAvailable() {
		databaseUnavailableJSON(w)
		return
	}

	userID, err := getUserIDFromCookie(r)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Not authenticated",
		})
		return
	}

	templateIDStr := r.URL.Query().Get("id")
	var templateID int64
	fmt.Sscanf(templateIDStr, "%d", &templateID)

	template, err := database.GetTemplateByID(templateID)
	if err != nil || template.ProfileID != userID {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Шаблон не найден",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"template": template,
	})
}
func UpdateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if !database.IsAvailable() {
		databaseUnavailableJSON(w)
		return
	}

	userID, err := getUserIDFromCookie(r)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Not authenticated",
		})
		return
	}

	var template domain.FormattingTemplate
	err = json.NewDecoder(r.Body).Decode(&template)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Некорректный запрос",
		})
		return
	}

	if err = validateTemplateName(&template); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	template.ProfileID = userID

	updatedTemplate, err := database.UpdateTemplate(&template)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"template": updatedTemplate,
	})
}

func validateTemplateName(template *domain.FormattingTemplate) error {
	template.Name = strings.TrimSpace(template.Name)
	if template.Name == "" {
		return fmt.Errorf("Название шаблона не может быть пустым")
	}
	if utf8.RuneCountInString(template.Name) > maxTemplateNameLength {
		return fmt.Errorf("Название шаблона не должно превышать %d символов", maxTemplateNameLength)
	}
	return nil
}
func DeleteTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if !database.IsAvailable() {
		databaseUnavailableJSON(w)
		return
	}

	userID, err := getUserIDFromCookie(r)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Not authenticated",
		})
		return
	}

	templateIDStr := r.URL.Query().Get("id")
	var templateID int64
	fmt.Sscanf(templateIDStr, "%d", &templateID)

	err = database.DeleteTemplate(templateID, userID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
func SelectTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if !database.IsAvailable() {
		databaseUnavailableJSON(w)
		return
	}

	userID, err := getUserIDFromCookie(r)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Not authenticated",
		})
		return
	}

	templateIDStr := r.URL.Query().Get("id")
	var templateID int64
	fmt.Sscanf(templateIDStr, "%d", &templateID)
	template, err := database.GetTemplateByID(templateID)
	if err != nil || template.ProfileID != userID {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Шаблон не найден",
		})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "selected_template",
		Value:    fmt.Sprintf("%d", templateID),
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30,
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"template": template,
		"message":  "Шаблон выбран",
	})
}
func ResetTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	http.SetCookie(w, &http.Cookie{
		Name:     "selected_template",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteLaxMode,
	})

	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
