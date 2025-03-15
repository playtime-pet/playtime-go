package handlers

import (
	"encoding/json"
	"net/http"
	"playtime-go/services"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	code := query.Get("code")

	if code == "" {
		http.Error(w, "Code is required", http.StatusBadRequest)
		return
	}

	session, err := services.GetLoginSession(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)

}
