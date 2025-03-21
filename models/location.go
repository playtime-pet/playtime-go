package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GeoLocation represents a geolocation point with coordinates
type GeoLocation struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

// Location represents a stored location in the system
type Location struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Address     string             `json:"address" bson:"address"`
	Description string             `json:"description" bson:"description"`
	Category    string             `json:"category" bson:"category"`
	Tags        []string           `json:"tags" bson:"tags"`
	Location    GeoLocation        `json:"location" bson:"location"` // GeoJSON format for MongoDB
	Phone       string             `json:"phone" bson:"phone"`
	Website     string             `json:"website" bson:"website"`
	PhotoURLs   []string           `json:"photoUrls" bson:"photoUrls"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// LocationRequest represents the incoming request to create or update a location
type LocationRequest struct {
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	Phone       string   `json:"phone"`
	Website     string   `json:"website"`
	PhotoURLs   []string `json:"photoUrls"`
}

// SearchRequest represents a request to search for nearby locations
type SearchRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Keyword   string  `json:"keyword"`
	Radius    float64 `json:"radius"` // Search radius in meters, default 1000
	Limit     int64   `json:"limit"`  // Maximum number of results, default 10
	Category  string  `json:"category"`
}

// SearchResult wraps a Location with additional distance information
type SearchResult struct {
	Location Location `json:"location"`
	Distance float64  `json:"distance"` // Distance to the search point in meters
}

// ReverseGeocodeResponse represents the response from Tencent Maps API
type ReverseGeocodeResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
		Address       string `json:"address"`
		AddressFormat struct {
			Recommend string `json:"recommend"`
			Rough     string `json:"rough"`
		} `json:"address_format"`
		AddressComponent struct {
			Nation       string `json:"nation"`
			Province     string `json:"province"`
			City         string `json:"city"`
			District     string `json:"district"`
			Street       string `json:"street"`
			StreetNumber string `json:"street_number"`
		} `json:"address_component"`
		AdInfo struct {
			Nation      string `json:"nation_code"`
			Province    string `json:"adcode"`
			City        string `json:"city_code"`
			District    string `json:"district_code"`
			Nationality string `json:"nationality_code"`
		} `json:"ad_info"`
		FormattedAddresses struct {
			Recommend string `json:"recommend"`
			Rough     string `json:"rough"`
		} `json:"formatted_addresses"`
		PoiCount int `json:"poi_count"`
		Pois     []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Address string `json:"address"`
			Category string `json:"category"`
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			Distance float64 `json:"_distance"`
		} `json:"pois"`
	} `json:"result"`
}