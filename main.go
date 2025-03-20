package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"playtime-go/db"
	"playtime-go/handlers"
	"playtime-go/services"
	"syscall"
)

func main() {
	// Initialize router
	router := http.NewServeMux()

	// Register routes
	router.HandleFunc("/token", handlers.HandleToken)
	router.HandleFunc("/phone", handlers.HandlePhone)
	router.HandleFunc("/user", handlers.HandleUser)
	router.HandleFunc("/user/", handlers.HandleUser) // This will catch all /user/* paths
	router.HandleFunc("/user/openid/", handlers.HandleUserByOpenID)
	router.HandleFunc("/wechat/", handlers.HandleWechat)
	router.HandleFunc("/pet", handlers.HandlePet)
	router.HandleFunc("/pet/", handlers.HandlePet) // This will catch all /pet/* paths
	router.HandleFunc("/map", handlers.HandleMap)
	router.HandleFunc("/map/", handlers.HandleMap)
	router.HandleFunc("/map/search", handlers.HandleMap)

	// Initialize MongoDB (connection is created on first use)
	db.GetMongoClient()

	// Create geospatial index for locations
	if err := services.EnsureLocationIndexes(); err != nil {
		log.Printf("Warning: Failed to create geospatial index: %v", err)
	}

	// Setup graceful shutdown
	setupGracefulShutdown()

	// Start the server
	fmt.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupGracefulShutdown registers handlers for SIGINT and SIGTERM signals
func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("Shutting down server...")

		// Clean up resources
		db.CloseMongoClient()

		fmt.Println("Server gracefully stopped")
		os.Exit(0)
	}()
}
