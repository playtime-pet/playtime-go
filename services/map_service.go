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
	"playtime-go/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const locationCollection = "locations"

// CreateLocation creates a new location in the database
func CreateLocation(request models.LocationRequest) (*models.LocationResponse, error) {
	// Validate coordinates
	if request.Latitude == 0 || request.Longitude == 0 {
		return nil, fmt.Errorf("invalid coordinates: latitude and longitude must be provided")
	}

	// Create new location with GeoJSON point for MongoDB geospatial queries
	now := time.Now()

	// Prepare tags based on pet-friendly information
	// tags := generateTags(request)

	// Determine category based on zone if not explicitly provided
	geoLocation := utils.ToGeoJSONPoint(request.Latitude, request.Longitude)

	// Debug log the GeoJSON data
	fmt.Printf("Creating location with GeoJSON: %+v\n", geoLocation)

	location := models.Location{
		BaseLocation: models.BaseLocation{
			Name:             request.Name,
			Address:          request.Address,
			Description:      request.Description,
			Category:         request.Category,
			IsPetFriendly:    request.IsPetFriendly,
			PetSize:          request.PetSize,
			PetType:          request.PetType,
			Zone:             request.Zone,
			AddressComponent: request.AddressComponent,
			AdInfo:           request.AdInfo,
		},
		Location:  geoLocation,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Insert location into database
	id, err := InsertOne(locationCollection, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create location: %v", err)
	}

	location.ID = id
	result, _ := ConvertLocationToResponse(location)
	if result == nil {
		return nil, fmt.Errorf("failed to convert location to response")
	}
	return result, nil
}

// Helper function to generate tags from pet-friendly information
// func generateTags(request models.LocationRequest) []string {
// 	tags := []string{}
// 	if request.IsPetFriendly {
// 		tags = append(tags, "pet-friendly")
// 		if request.PetSize != "" {
// 			tags = append(tags, "pet-size-"+request.PetSize)
// 		}
// 		if request.PetType != "" {
// 			tags = append(tags, "pet-type-"+request.PetType)
// 		}
// 	}
// 	return tags
// }

// GetLocationByID retrieves a location by ID
func GetLocationByID(id primitive.ObjectID) (*models.LocationResponse, error) {
	filter := bson.M{"_id": id}
	var location models.Location

	err := FindOne(locationCollection, filter, &location)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no location found with ID: %s", id.Hex())
		}
		return nil, fmt.Errorf("failed to get location by ID: %v", err)
	}

	result, _ := ConvertLocationToResponse(location)
	if result == nil {
		return nil, fmt.Errorf("failed to convert location to response")
	}
	return result, nil
}

func ConvertLocationToResponse(location models.Location) (*models.LocationResponse, error) {
	if location.Location.Type == "" {
		return nil, fmt.Errorf("invalid location: missing GeoJSON type")
	}

	if len(location.Location.Coordinates) == 0 {
		return nil, fmt.Errorf("invalid location: missing coordinates")
	}

	latitude, longitude, err := utils.FromGeoJSONPoint(location.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to convert GeoJSON coordinates: %v", err)
	}

	response := &models.LocationResponse{
		ID:           location.ID,
		BaseLocation: location.BaseLocation,
		Latitude:     latitude,
		Longitude:    longitude,
	}

	fmt.Printf("Converted location: %+v\n", response)
	return response, nil
}

// UpdateLocation updates an existing location
func UpdateLocation(id primitive.ObjectID, request models.LocationRequest) (*models.LocationResponse, error) {
	// Check if location exists
	existing, err := GetLocationByID(id)
	if err != nil {
		return nil, err
	}

	geoLocation := utils.ToGeoJSONPoint(existing.Latitude, existing.Longitude)

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
			"addressComponent": existing.AddressComponent,
			"adInfo":           existing.AdInfo,
			"category":         request.Zone,
			"location":         geoLocation,
			"updatedAt":        now,
		},
	}

	// Update location in the database
	filter := bson.M{"_id": id}
	err = UpdateOne(locationCollection, filter, updateData)
	if err != nil {
		return nil, fmt.Errorf("failed to update location: %v", err)
	}

	// Get the updated location
	result, _ := GetLocationByID(id)
	if result == nil {
		return nil, fmt.Errorf("failed to get updated location")
	}
	return result, nil
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
func ListLocations(category string, limit int64) ([]models.LocationResponse, error) {
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

	// Fetch locations from database
	err := FindMany(locationCollection, filter, &locations, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %v", err)
	}

	// Handle empty result case
	if len(locations) == 0 {
		return []models.LocationResponse{}, nil
	}

	// Convert locations to response format
	responses := make([]models.LocationResponse, 0, len(locations))
	for _, location := range locations {
		jsonLocation, _ := json.MarshalIndent(location, "", "  ")
		fmt.Println(string(jsonLocation))

		response, err := ConvertLocationToResponse(location)
		fmt.Println(response)
		if err != nil {
			// Log the error but continue with other locations
			// This prevents a single conversion error from failing the entire request
			fmt.Printf("failed to convert location %s: %v\n", location.ID.Hex(), err)
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
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

	// Create the $geoNear pipeline stage - Remove limit parameter from here
	geoNearStage := bson.D{
		{Key: "$geoNear", Value: bson.D{
			{Key: "near", Value: bson.D{
				{Key: "type", Value: "Point"},
				{Key: "coordinates", Value: []float64{search.Longitude, search.Latitude}},
			}},
			{Key: "distanceField", Value: "distance"},
			{Key: "maxDistance", Value: radius},
			{Key: "spherical", Value: true},
			// Removed: {Key: "limit", Value: limit},
		}},
	}

	// Initialize pipeline with geoNear stage
	pipeline := []bson.D{geoNearStage}

	// Add filtering by keyword if provided
	if search.Keyword != "" {
		// Text search across multiple fields
		matchStage := bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "$or", Value: []bson.D{
					{{Key: "name", Value: bson.D{{Key: "$regex", Value: search.Keyword}, {Key: "$options", Value: "i"}}}},
					{{Key: "description", Value: bson.D{{Key: "$regex", Value: search.Keyword}, {Key: "$options", Value: "i"}}}},
					// Keep if you have tags field, comment out if not
					// {{Key: "tags", Value: bson.D{{Key: "$regex", Value: search.Keyword}, {Key: "$options", Value: "i"}}}},
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

	// Add limit stage at the end of the pipeline
	limitStage := bson.D{{Key: "$limit", Value: limit}}
	pipeline = append(pipeline, limitStage)

	// Execute the aggregation
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to execute nearby search: %v", err)
	}
	defer cursor.Close(ctx)

	// Process results - this is where the fix is needed
	var results []models.SearchResult

	// Use a map to hold each document with its distance field
	for cursor.Next(ctx) {
		// Use a raw document to decode
		var rawDoc bson.M
		if err := cursor.Decode(&rawDoc); err != nil {
			return nil, fmt.Errorf("failed to decode search result: %v", err)
		}

		// Extract the distance
		distance, _ := rawDoc["distance"].(float64)
		fmt.Print(rawDoc)
		// Delete the distance field from the document
		delete(rawDoc, "distance")

		// Extract coordinates from location
		if location, ok := rawDoc["location"].(bson.M); ok {
			if coords, ok := location["coordinates"].(primitive.A); ok && len(coords) == 2 {
				longitude := coords[0].(float64)
				latitude := coords[1].(float64)
				// Add coordinates to rawDoc for later use
				rawDoc["longitude"] = longitude
				rawDoc["latitude"] = latitude
			}
		}

		// Marshal and unmarshal to convert to a Location object
		bsonData, err := bson.Marshal(rawDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal document: %v", err)
		}

		var location models.Location
		if err := bson.Unmarshal(bsonData, &location); err != nil {
			return nil, fmt.Errorf("failed to unmarshal to location: %v", err)
		}

		convertLocation, _ := ConvertLocationToResponse(location)
		if convertLocation == nil {
			return nil, fmt.Errorf("failed to convert location to response")
		}
		// Add to results
		results = append(results, models.SearchResult{
			Location: *convertLocation,
			Distance: distance,
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
	params.Add("get_poi", "1") // Get nearby POIs

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
