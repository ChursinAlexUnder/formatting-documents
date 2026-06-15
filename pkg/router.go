package pkg

import (
	"formatting-documents/internal/config"
	"formatting-documents/internal/interfaces"
	"net/http"
)

func ConnectionStatic() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(config.RootPath("web", "static")))))
}

func HandlerPages() {
	http.HandleFunc("/", interfaces.MainPage)
	http.HandleFunc("/menu", interfaces.ShowOptions)
	http.HandleFunc("/download", interfaces.SendDocument)
	http.HandleFunc("/error", interfaces.ErrorPage)
	http.HandleFunc("/errortime", interfaces.ErrorTimePage)
	http.HandleFunc("/events", interfaces.SSEChannel)
	http.HandleFunc("/info", interfaces.InfoPage)
	http.HandleFunc("/profile", interfaces.ProfilePage)
	http.HandleFunc("/api/auth/register", interfaces.RegisterHandler)
	http.HandleFunc("/api/auth/login", interfaces.LoginHandler)
	http.HandleFunc("/api/auth/logout", interfaces.LogoutHandler)
	http.HandleFunc("/api/config/turnstile", interfaces.TurnstileConfigHandler)
	http.HandleFunc("/api/profile", interfaces.GetProfileHandler)
	http.HandleFunc("/api/templates/create", interfaces.CreateTemplateHandler)
	http.HandleFunc("/api/templates/get", interfaces.GetTemplateHandler)
	http.HandleFunc("/api/templates/update", interfaces.UpdateTemplateHandler)
	http.HandleFunc("/api/templates/delete", interfaces.DeleteTemplateHandler)
	http.HandleFunc("/api/templates/select", interfaces.SelectTemplateHandler)
	http.HandleFunc("/api/templates/reset", interfaces.ResetTemplateHandler)
}
