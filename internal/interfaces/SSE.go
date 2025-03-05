package interfaces

import (
	"encoding/json"
	"fmt"
	"formatting-documents/internal/domain"
	"formatting-documents/internal/services"
	"net/http"
	"time"
)

func SSEChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Fprint(w, "Error: streaming unsupported")
		return
	}

	var prevResponse string

	// Создаем канал для отслеживания закрытия соединения
	closeClient := r.Context().Done()

	for {
		select {
		case <-closeClient:
			fmt.Fprint(w, "Client disconnected")
			return
		default:
			data, err := services.ReadFileJSON("../data.json")
			if err != nil {
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
				return
			}

			response := domain.SSEData{
				Count:          data.Count,
				LastFormatting: data.LastFormatting,
			}

			jsonResponse, err := json.Marshal(response)
			if err != nil {
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
				return
			}

			currentResponse := string(jsonResponse)
			if prevResponse != currentResponse {
				fmt.Fprintf(w, "data: %s\n\n", currentResponse)
				flusher.Flush()
				prevResponse = currentResponse
			}

			time.Sleep(time.Second)
		}
	}
}
