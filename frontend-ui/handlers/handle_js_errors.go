package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

//map the incoming JSON error string from js
type ErrorLogRequest struct {
	Error string `json:"errerror"`
}

func (c *BackendClient) ErrorLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Fatal("POST required, got:",r.Method)
	}

	var req ErrorLogRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Fatal("Error decoding json: ", err)
	}

	// For now, prints directly to your Go system terminal logs
	log.Printf("[Frontend JS Error] 🚨 Happened at %s: %s\n", time.Now().Format("15:04:05"), req.Error)

	// Return a clean acknowledgment status back to JS
	w.WriteHeader(http.StatusNoContent)

}