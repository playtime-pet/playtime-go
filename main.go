package main

import (
	"fmt"
	"log"
	"net/http"
	"playtime-go/handlers"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/token", handlers.HandleToken)
	router.HandleFunc("/phone", handlers.HandlePhone)
	
	fmt.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
