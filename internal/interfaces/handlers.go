package interfaces

import (
	"encoding/json"
	"fmt"
	"formatting-documents/database"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/infrastructure"
	"formatting-documents/internal/services"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"text/template"
	"time"
)

func getUserIDFromCookie(r *http.Request) (int64, error) {
	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		return 0, fmt.Errorf("not authenticated")
	}

	var userID int64
	if _, err := fmt.Sscanf(cookie.Value, "%d", &userID); err != nil || userID <= 0 {
		return 0, fmt.Errorf("invalid session")
	}

	return userID, nil
}

func databaseUnavailableJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": "Database is unavailable",
	})
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	var (
		wrongData domain.WrongData = domain.WrongData{}
		data      domain.Answer
		err       error
	)
	err = services.CheckDataJSON()
	if err != nil {
		fmt.Fprintf(w, "Error checking JSON data: %v", err)
		return
	}
	if r.Method == http.MethodPost {
		data, wrongData, err = ManagementData(w, r)
		if err != nil && err.Error() != "error validation" {
			fmt.Fprintf(w, "Error managementing: %v", err)
			return
		} else if err == nil {
			SendDocumentPage(w, r, data)
			return
		}
	}
	// путь откуда вызываю эту функцию, а не от её расположения
	tmplt, err := template.ParseFiles("../web/templates/index.html", "../web/templates/main.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing index.html and main.html: %v", err)
		return
	}
	// отображение страницы в заготовленном каркасе с footer
	err = tmplt.ExecuteTemplate(w, "index", wrongData)
	if err != nil {
		fmt.Fprintf(w, "Error displaying index.html and main.html: %v", err)
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

	// Возвращаем данные в формате JSON
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"options": options,
	})
	if err != nil {
		fmt.Fprintf(w, "Error sending json response: %v", err)
		return
	}
}

func SendDocumentPage(w http.ResponseWriter, r *http.Request, data domain.Answer) {
	var (
		fullData      domain.AnswerWithInterfaceName
		interfaceName string
	)

	// Загружаем и парсим шаблоны, добавляя пользовательскую функцию add
	tmplt, err := template.New("index.html").Funcs(template.FuncMap{
		"add": services.Add,
	}).ParseFiles("../web/templates/index.html", "../web/templates/download.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing download.html: %v", err)
		return
	}

	data.DocumentData.Filename = "formatted_" + data.DocumentData.Filename
	interfaceName = data.DocumentData.Filename[:10] + data.DocumentData.Filename[15:]
	// кодирование для URL
	data.DocumentData.Filename = url.QueryEscape(data.DocumentData.Filename)
	fullData = domain.AnswerWithInterfaceName{Data: data, InterfaceName: interfaceName}
	err = tmplt.ExecuteTemplate(w, "index", fullData)
	if err != nil {
		fmt.Fprintf(w, "Error displaying index.html and download.html: %v", err)
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
		fmt.Fprint(w, "Error getting the name of a new document from the page")
		return
	}
	// декодируем
	formattedDocumentName, err := url.QueryUnescape(formattedDocumentName)
	if err != nil {
		fmt.Fprintf(w, "Error decoding the name of a new document: %v", err)
		return
	}
	formattedDocumentPath = "../buffer/" + formattedDocumentName

	// проверка на наличие документа
	if _, err := os.Stat(formattedDocumentPath); err != nil {
		// перенаправление на страницу с ошибкой
		http.Redirect(w, r, "/error", http.StatusSeeOther)
		return
	}

	formattedDocument, err = os.Open(formattedDocumentPath)
	if err != nil {
		fmt.Fprintf(w, "Error opening the new document: %v", err)
		return
	}
	defer func() {
		formattedDocument.Close()
		infrastructure.DeleteBothDocuments(formattedDocumentName)
	}()

	// проверка на проблемы в файле
	formattedDocumentInfo, err := formattedDocument.Stat()
	if err != nil {
		fmt.Fprintf(w, "Error: problem in the document: %v", err)
		return
	}

	interfaceName = formattedDocumentName[:10] + formattedDocumentName[15:]

	// установка заголовков
	w.Header().Set("Content-Disposition", "attachment; filename="+interfaceName)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Length", strconv.Itoa(int(formattedDocumentInfo.Size())))

	// отдача файла пользователю
	_, err = io.Copy(w, formattedDocument)
	if err != nil {
		fmt.Fprintf(w, "Error sending new document: %v", err)
		return
	}
}

func ErrorPage(w http.ResponseWriter, r *http.Request) {
	tmplt, err := template.ParseFiles("../web/templates/index.html", "../web/templates/error.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing error.html: %v", err)
		return
	}
	err = tmplt.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Fprintf(w, "Error displaying index.html and error.html: %v", err)
		return
	}
}

func ErrorTimePage(w http.ResponseWriter, r *http.Request) {
	tmplt, err := template.ParseFiles("../web/templates/index.html", "../web/templates/errortime.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing errortime.html: %v", err)
		return
	}
	err = tmplt.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Fprintf(w, "Error displaying index.html and errortime.html: %v", err)
		return
	}
}

func InfoPage(w http.ResponseWriter, r *http.Request) {
	tmplt, err := template.ParseFiles("../web/templates/index.html", "../web/templates/info.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing info.html: %v", err)
		return
	}
	err = tmplt.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Fprintf(w, "Error displaying index.html and info.html: %v", err)
		return
	}
}

// Profile page handler
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

	tmplt, err := template.ParseFiles("../web/templates/index.html", "../web/templates/profile.html")
	if err != nil {
		fmt.Fprintf(w, "Error parsing profile.html: %v", err)
		return
	}
	err = tmplt.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Fprintf(w, "Error displaying index.html and profile.html: %v", err)
		return
	}
}

// API: Register handler
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
			"message": "Invalid request",
		})
		return
	}

	if len(req.Login) == 0 || len(req.Password) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
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

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    fmt.Sprintf("%d", user.ID),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7, // 7 days
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Registration successful",
		"user_id": user.ID,
	})
}

// API: Login handler
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
			"message": "Invalid request",
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

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    fmt.Sprintf("%d", userID),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7, // 7 days
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"user_id": userID,
	})
}

// API: Logout handler
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

// API: Get user profile with templates
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
			"message": "User not found",
		})
		return
	}

	templates, err := database.GetTemplatesByUserID(userID)
	if err != nil {
		templates = []domain.FormattingTemplate{}
	}

	// Get selected template ID from cookie
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

// API: Create template
func CreateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
			"message": "Invalid request",
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

// API: Get template
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
			"message": "Template not found",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"template": template,
	})
}

// API: Update template
func UpdateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
			"message": "Invalid request",
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

// API: Delete template
func DeleteTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

// API: Select template
func SelectTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	// Verify template belongs to user
	template, err := database.GetTemplateByID(templateID)
	if err != nil || template.ProfileID != userID {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Template not found",
		})
		return
	}

	// Set cookie with selected template
	http.SetCookie(w, &http.Cookie{
		Name:     "selected_template",
		Value:    fmt.Sprintf("%d", templateID),
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30, // 30 days
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"template": template,
		"message":  "Template selected",
	})
}

// API: Reset template selection
func ResetTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
