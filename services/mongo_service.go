package services

import (
	"context"
	"fmt"
	"playtime-go/db"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	userCollection = "users"
	timeout        = 10 * time.Second
)

// InsertOne inserts a document into the specified collection
func InsertOne(collectionName string, document interface{}) (primitive.ObjectID, error) {
	collection := db.GetCollection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to insert document: %v", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid, nil
	}

	return primitive.NilObjectID, fmt.Errorf("failed to get inserted ID")
}

// FindOne finds a single document matching the filter in the specified collection
func FindOne(collectionName string, filter interface{}, result interface{}) error {
	collection := db.GetCollection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return err
		}
		return fmt.Errorf("failed to find document: %v", err)
	}

	return nil
}

// FindMany finds multiple documents matching the filter in the specified collection
func FindMany(collectionName string, filter interface{}, result interface{}, opts ...*options.FindOptions) error {
	collection := db.GetCollection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, opts...)
	if err != nil {
		return fmt.Errorf("failed to find documents: %v", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, result); err != nil {
		return fmt.Errorf("failed to decode documents: %v", err)
	}

	return nil
}

// UpdateOne updates a single document matching the filter in the specified collection
func UpdateOne(collectionName string, filter interface{}, update interface{}) error {
	collection := db.GetCollection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update document: %v", err)
	}

	return nil
}

// DeleteOne deletes a single document matching the filter in the specified collection
func DeleteOne(collectionName string, filter interface{}) error {
	collection := db.GetCollection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete document: %v", err)
	}

	return nil
}

// Count counts the number of documents matching the filter in the specified collection
func Count(collectionName string, filter interface{}) (int64, error) {
	collection := db.GetCollection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %v", err)
	}

	return count, nil
}
