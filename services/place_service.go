package services

import (
	"context"
	"fmt"
	"playtime-go/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const reviewCollection = "reviews"

// CreateReview creates a new review in the database
func CreateReview(request models.Review) (*models.Review, error) {
	// Set current time as review date
	now := time.Now()
	request.Date = now

	// Insert review into database
	_, err := InsertOne(reviewCollection, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create review: %v", err)
	}

	return &request, nil
}

// GetReview retrieves a review by ID
func GetReview(id primitive.ObjectID) (*models.Review, error) {
	filter := bson.M{"_id": id}
	var review models.Review

	err := FindOne(reviewCollection, filter, &review)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no review found with ID: %s", id.Hex())
		}
		return nil, fmt.Errorf("failed to get review by ID: %v", err)
	}

	return &review, nil
}

// UpdateReview updates an existing review
func UpdateReview(id primitive.ObjectID, request models.Review) (*models.Review, error) {
	// Check if review exists
	_, err := GetReview(id)
	if err != nil {
		return nil, err
	}

	// Prepare update document
	updateData := bson.M{
		"$set": bson.M{
			"content":    request.Content,
			"ratingStar": request.Rating,
			"date":       time.Now(),
		},
	}

	// Update review in the database
	filter := bson.M{"_id": id}
	err = UpdateOne(reviewCollection, filter, updateData)
	if err != nil {
		return nil, fmt.Errorf("failed to update review: %v", err)
	}

	// Get the updated review
	return GetReview(id)
}

// DeleteReview deletes a review by ID
func DeleteReview(id primitive.ObjectID) error {
	// Check if review exists
	_, err := GetReview(id)
	if err != nil {
		return err
	}

	// Delete review from the database
	filter := bson.M{"_id": id}
	err = DeleteOne(reviewCollection, filter)
	if err != nil {
		return fmt.Errorf("failed to delete review: %v", err)
	}

	return nil
}

// GetReviewsByPlace gets all reviews for a specific place
func GetReviewsByPlace(placeID string, limit int64) ([]models.Review, error) {
	filter := bson.M{"placeId": placeID}
	var reviews []models.Review

	// Set options for sorting by date (descending) and limit
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "date", Value: -1}})

	// Apply limit if specified
	if limit > 0 {
		findOptions.SetLimit(limit)
	} else {
		findOptions.SetLimit(100) // Default limit
	}

	err := FindMany(reviewCollection, filter, &reviews, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews by place: %v", err)
	}

	return reviews, nil
}

// GetReviewsByUserID gets all reviews for a specific user
func GetReviewsByUserID(ctx context.Context, userID string) ([]models.Review, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %v", err)
	}

	filter := bson.M{"userId": userID}
	var reviews []models.Review

	// Set options for sorting by date (descending)
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "date", Value: -1}})

	err := FindMany(reviewCollection, filter, &reviews, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews by user: %v", err)
	}

	return reviews, nil
}
