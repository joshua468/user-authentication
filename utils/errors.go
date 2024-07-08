package utils

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := ErrorResponse{Message: message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}
