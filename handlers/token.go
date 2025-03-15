package handlers

import (
	"net/http"
	"playtime-go/services"
	"playtime-go/utils"
)

func HandleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", 405, http.StatusMethodNotAllowed)
		return
	}

	token, err := services.GetToken()
	if err != nil {
		utils.ErrorResponse(w, err.Error(), 500, http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, token, http.StatusOK)
}
