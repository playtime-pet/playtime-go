package utils

import (
	"log"
	"net/http"
)

// LoggingMiddleware wraps an http.HandlerFunc with request logging
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := GenerateRequestID()
		log.Printf("[%s] Started %s request to %s", requestID, r.Method, r.URL.Path)

		defer func() {
			if err := recover(); err != nil {
				log.Printf("[%s] Panic occurred: %v", requestID, err)
				ErrorResponse(w, "Internal server error", 500, http.StatusInternalServerError)
			}
			log.Printf("[%s] Completed %s request to %s", requestID, r.Method, r.URL.Path)
		}()

		next(w, r)
	}
}
