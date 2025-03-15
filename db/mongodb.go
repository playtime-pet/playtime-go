package db

import (
	"context"
	"fmt"
	"log"
	"playtime-go/config"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	database   *mongo.Database
	clientOnce sync.Once
)

// GetMongoClient returns a singleton MongoDB client
func GetMongoClient() *mongo.Client {
	clientOnce.Do(func() {
		cfg := config.GetConfig()
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.MongoTimeout)*time.Second)
		defer cancel()

		// Create client options
		clientOptions := options.Client().ApplyURI(cfg.MongoURI)

		fmt.Println(cfg)
		// Add credentials if username and password are provided
		if cfg.MongoUser != "" && cfg.MongoPass != "" {
			credential := options.Credential{
				Username: cfg.MongoUser,
				Password: cfg.MongoPass,
			}
			clientOptions.SetAuth(credential)
		}

		// Connect to MongoDB
		var err error
		client, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		// Ping the MongoDB server to verify the connection
		err = client.Ping(ctx, nil)
		if err != nil {
			log.Fatalf("Failed to ping MongoDB: %v", err)
		}

		log.Println("Connected to MongoDB successfully")
		database = client.Database(cfg.MongoDB)
	})

	return client
}

// GetCollection returns a MongoDB collection
func GetCollection(collectionName string) *mongo.Collection {
	GetMongoClient() // Ensure the client is initialized
	return database.Collection(collectionName)
}

// CloseMongoClient closes the MongoDB client connection
func CloseMongoClient() {
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error closing MongoDB connection: %v", err)
		}
		log.Println("MongoDB connection closed")
	}
}
