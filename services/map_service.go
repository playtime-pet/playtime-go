package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"playtime-go/config"
	"playtime-go/db"
	"playtime-go/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const locationCollection = "locations"

// CreateLocation creates a new location in the database
func CreateLocation(request models.LocationRequest) (*models.Location, error) {
	// Create new location with GeoJSON point for MongoDB geospatial queries
	now := time.Now()
	
	// Prepare tags based on pet-friendly information
	// tags := generateTags(request)
	
	// Determine category based on zone if not explicitly provided
	category := request.Zone
	
	location := models.Location{
		Name:             request.Name,
		Address:          request.Address,
		Description:      request.Description,
		Category:         category,
		// Tags:             tags,
		Location: models.GeoLocation{
			Type:        "Point",
			Coordinates: []float64{request.Longitude, request.Latitude}, // GeoJSON format: [longitude, latitude]
		},
		IsPetFriendly:    request.IsPetFriendly,
		PetSize:          request.PetSize,
		PetType:          request.PetType,
		Zone:             request.Zone,
		AddressComponent: request.AddressComponent,
		AdInfo:           request.AdInfo,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Insert location into database
	id, err := InsertOne(locationCollection, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create location: %v", err)
	}

	location.ID = id
	return &location, nil
}

// Helper function to generate tags from pet-friendly information
func generateTags(request models.LocationRequest) []string {
	tags := []string{}
	if request.IsPetFriendly {
		tags = append(tags, "pet-friendly")
		if request.PetSize != "" {
			tags = append(tags, "pet-size-"+request.PetSize)
		}
		if request.PetType != "" {
			tags = append(tags, "pet-type-"+request.PetType)
		}
	}
	return tags
}

// GetLocationByID retrieves a location by ID
func GetLocationByID(id primitive.ObjectID) (*models.Location, error) {
	filter := bson.M{"_id": id}
	var location models.Location

	err := FindOne(locationCollection, filter, &location)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no location found with ID: %s", id.Hex())
		}
		return nil, fmt.Errorf("failed to get location by ID: %v", err)
	}

	return &location, nil
}

// UpdateLocation updates an existing location
func UpdateLocation(id primitive.ObjectID, request models.LocationRequest) (*models.Location, error) {
	// Check if location exists
	_, err := GetLocationByID(id)
	if err != nil {
		return nil, err
	}

	// Prepare update document with GeoJSON point
	now := time.Now()
	updateData := bson.M{
		"$set": bson.M{
			"name":             request.Name,
			"address":          request.Address,
			"description":      request.Description,
			"isPetFriendly":    request.IsPetFriendly,
			"petSize":          request.PetSize,
			"petType":          request.PetType,
			"zone":             request.Zone,
			"addressComponent": request.AddressComponent,
			"adInfo":           request.AdInfo,
			"category":         request.Zone,
			// "tags":             generateTags(request),
			"location": bson.M{
				"type":        "Point",
				"coordinates": []float64{request.Longitude, request.Latitude},
			},
			"updatedAt": now,
		},
	}

	// Update location in the database
	filter := bson.M{"_id": id}
	err = UpdateOne(locationCollection, filter, updateData)
	if err != nil {
		return nil, fmt.Errorf("failed to update location: %v", err)
	}

	// Get the updated location
	return GetLocationByID(id)
}

// DeleteLocation deletes a location by ID
func DeleteLocation(id primitive.ObjectID) error {
	// Check if location exists
	_, err := GetLocationByID(id)
	if err != nil {
		return err
	}

	// Delete location from the database
	filter := bson.M{"_id": id}
	err = DeleteOne(locationCollection, filter)
	if err != nil {
		return fmt.Errorf("failed to delete location: %v", err)
	}

	return nil
}

// ListLocations retrieves all locations with optional filtering
func ListLocations(category string, limit int64) ([]models.Location, error) {
	// Prepare filter
	filter := bson.M{}
	if category != "" {
		filter["category"] = category
	}

	var locations []models.Location

	// Set options for sorting and limit
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "name", Value: 1}})

	// Apply limit if specified
	if limit > 0 {
		findOptions.SetLimit(limit)
	} else {
		findOptions.SetLimit(100) // Default limit
	}

	err := FindMany(locationCollection, filter, &locations, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %v", err)
	}

	return locations, nil
}

// SearchNearbyLocations searches for locations near the specified coordinates
func SearchNearbyLocations(search models.SearchRequest) ([]models.SearchResult, error) {
	collection := db.GetCollection(locationCollection)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Set default radius if not specified
	radius := search.Radius
	if radius <= 0 {
		radius = 1000 // Default radius: 1000 meters (1km)
	}

	// Set default limit if not specified
	limit := search.Limit
	if limit <= 0 {
		limit = 10 // Default limit: 10 results
	}

	// Create the $geoNear pipeline stage
	geoNearStage := bson.D{
		{Key: "$geoNear", Value: bson.D{
			{Key: "near", Value: bson.D{
				{Key: "type", Value: "Point"},
				{Key: "coordinates", Value: []float64{search.Longitude, search.Latitude}},
			}},
			{Key: "distanceField", Value: "distance"},
			{Key: "maxDistance", Value: radius},
			{Key: "spherical", Value: true},
			{Key: "limit", Value: limit},
		}},
	}

	// Add filtering by keyword if provided
	pipeline := []bson.D{geoNearStage}
	
	if search.Keyword != "" {
		// Text search across multiple fields
		matchStage := bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "$or", Value: []bson.D{
					{{Key: "name", Value: bson.D{{Key: "$regex", Value: search.Keyword}, {Key: "$options", Value: "i"}}}},
					{{Key: "description", Value: bson.D{{Key: "$regex", Value: search.Keyword}, {Key: "$options", Value: "i"}}}},
					{{Key: "tags", Value: bson.D{{Key: "$regex", Value: search.Keyword}, {Key: "$options", Value: "i"}}}},
				}},
			}},
		}
		pipeline = append(pipeline, matchStage)
	}
	
	// Add category filter if provided
	if search.Category != "" {
		categoryStage := bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "category", Value: search.Category},
			}},
		}
		pipeline = append(pipeline, categoryStage)
	}

	// Execute the aggregation
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to execute nearby search: %v", err)
	}
	defer cursor.Close(ctx)

	// Structure to decode the aggregation results
	type aggregateResult struct {
		models.Location
		Distance float64 `bson:"distance"`
	}

	// Process results
	var results []models.SearchResult
	var aggResult aggregateResult
	
	for cursor.Next(ctx) {
		if err := cursor.Decode(&aggResult); err != nil {
			return nil, fmt.Errorf("failed to decode search result: %v", err)
		}
		
		results = append(results, models.SearchResult{
			Location: aggResult.Location,
			Distance: aggResult.Distance,
		})
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return results, nil
}

// EnsureLocationIndexes creates the necessary geospatial indexes for location queries
func EnsureLocationIndexes() error {
	collection := db.GetCollection(locationCollection)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a 2dsphere index on the location field
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
		Options: options.Index().
			SetName("location_2dsphere").
			SetBackground(true),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create geospatial index: %v", err)
	}

	return nil
}

// ReverseGeocode calls Tencent Maps API to convert lat/lng to an address
func ReverseGeocode(lat string, lng string) (interface{}, error) {
	// Get the API key from config
	cfg := config.GetConfig()
	if cfg.MiniMapKey == "" {
		return nil, fmt.Errorf("tencent map API key is not configured")
	}

	// Build request URL with parameters
	baseURL := "https://apis.map.qq.com/ws/geocoder/v1/"
	params := url.Values{}
	params.Add("key", cfg.MiniMapKey)
	params.Add("location", fmt.Sprintf("%s,%s", lat, lng))
	params.Add("get_poi", "1")  // Get nearby POIs
	
	// Build the full URL with parameters
	apiURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	
	// Make HTTP request to Tencent Maps API
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call Tencent Maps API: %v", err)
	}
	defer resp.Body.Close()
	
	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tencent maps API returned non-200 status code: %d", resp.StatusCode)
	}
	
	// Parse the response
	var geocodeResponse models.ReverseGeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&geocodeResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Tencent Maps API response: %v", err)
	}
	
	// Check if the API returned an error
	if geocodeResponse.Status != 0 {
		return nil, fmt.Errorf("tencent maps API error: %d - %s", geocodeResponse.Status, geocodeResponse.Message)
	}
	
	// Return the result
	return geocodeResponse.Result, nil
}