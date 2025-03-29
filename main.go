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
	"playtime-go/utils"
	"syscall"
)

func main() {
	// Initialize router
	router := http.NewServeMux()

	// Register routes with logging middleware
	router.HandleFunc("/token", utils.LoggingMiddleware(handlers.HandleToken))
	router.HandleFunc("/phone", utils.LoggingMiddleware(handlers.HandlePhone))
	router.HandleFunc("/wechat/", utils.LoggingMiddleware(handlers.HandleWechat))

	// User routes - explicitly handle both /user and /user/ patterns
	router.HandleFunc("/user/openid/", utils.LoggingMiddleware(handlers.HandleUserByOpenID))
	router.HandleFunc("/user", utils.LoggingMiddleware(handlers.HandleUser))  // Exact match for /user
	router.HandleFunc("/user/", utils.LoggingMiddleware(handlers.HandleUser)) // Prefix match for /user/123

	// pet related
	router.HandleFunc("/pet", utils.LoggingMiddleware(handlers.HandlePet))
	router.HandleFunc("/pet/", utils.LoggingMiddleware(handlers.HandlePet)) // This will catch all /pet/* paths

	router.HandleFunc("/place", utils.LoggingMiddleware(handlers.HandlePlace)) // This will catch all /place/* paths
	router.HandleFunc("/place/", utils.LoggingMiddleware(handlers.HandlePlace))

	// review related
	router.HandleFunc("/review/user/", utils.LoggingMiddleware(handlers.HandleReview))  // handle user reviews
	router.HandleFunc("/review/place/", utils.LoggingMiddleware(handlers.HandleReview)) // handler place reviews
	router.HandleFunc("/review/", utils.LoggingMiddleware(handlers.HandleReview))

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
