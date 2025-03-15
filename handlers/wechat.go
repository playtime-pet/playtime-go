package handlers

import (
	"net/http"
	"playtime-go/services"
	"playtime-go/utils"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", 400, http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	code := query.Get("code")

	if code == "" {
		utils.ErrorResponse(w, "Code is required", 401, http.StatusBadRequest)
		return
	}

	session, err := services.GetLoginSession(code)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, session, http.StatusOK)
}
