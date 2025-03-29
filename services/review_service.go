package services

import (
	"fmt"
	"playtime-go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DeleteAllUserReview(userID primitive.ObjectID) error {
	// Create filter to match all reviews by the user
	filter := bson.M{"user_id": userID}

	// Use the DeleteMany function from mongo_service.go
	deletedCount, err := DeleteMany(reviewCollection, filter)
	if err != nil {
		return fmt.Errorf("failed to delete user reviews: %v", err)
	}

	fmt.Printf("%v rows of user reviews are delete \n", deletedCount)
	// Log the number of deleted documents (optional)
	// log.Printf("Deleted %d reviews for user %s", deletedCount, userID.Hex())

	return nil
}

func GetAllUserReview(userID primitive.ObjectID) ([]models.Review, error) {
	filter := bson.M{"user_id": userID}
	var reviews []models.Review

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "date", Value: -1}})

	err := FindMany(reviewCollection, filter, &reviews, findOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no review found with ID: %s", userID.Hex())
		}
		return nil, fmt.Errorf("failed to get review by ID: %v", err)
	}

	return reviews, nil
}

func GetAllPlaceReview(placeID primitive.ObjectID) ([]models.Review, error) {
	filter := bson.M{"place_id": placeID}
	var reviews []models.Review

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "date", Value: -1}})

	err := FindMany(reviewCollection, filter, &reviews, findOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no review found for place ID: %s", placeID.Hex())
		}
		return nil, fmt.Errorf("failed to get reviews for place: %v", err)
	}

	return reviews, nil
}

func DeleteAllPlaceReview(placeID primitive.ObjectID) error {
	// Create filter to match all reviews for the place
	filter := bson.M{"place_id": placeID}

	// Use the DeleteMany function from mongo_service.go
	deletedCount, err := DeleteMany(reviewCollection, filter)
	if err != nil {
		return fmt.Errorf("failed to delete place reviews: %v", err)
	}

	fmt.Printf("%v rows of place reviews are deleted\n", deletedCount)

	return nil
}
