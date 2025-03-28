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

// BaseLocation contains common fields shared across location-related structs
type BaseLocation struct {
	Name             string           `json:"name" bson:"name" validate:"required,min=2,max=100"`
	Address          string           `json:"address" bson:"address" validate:"required,min=5,max=200"`
	Description      string           `json:"description" bson:"description" validate:"max=500"`
	Category         string           `json:"category" bson:"category" validate:"required,oneof=park cafe restaurant shop other"`
	Photos           []string         `json:"photos,omitempty" bson:"photos,omitempty" validate:"max=10"`
	IsPetFriendly    bool             `json:"isPetFriendly" bson:"isPetFriendly"`
	PetSize          []string         `json:"petSize" bson:"petSize" validate:"dive,omitempty,oneof=small medium large"`
	PetType          []string         `json:"petType" bson:"petType" validate:"dive,omitempty,oneof=dog cat other"`
	Zone             []string         `json:"zone" bson:"zone" validate:"required,dive,required"`
	AddressComponent AddressComponent `json:"addressComponent" bson:"addressComponent" validate:"required"`
	AdInfo           AdInfo           `json:"adInfo" bson:"adInfo" validate:"required"`
}

// Location represents a stored location in the system
type Location struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	BaseLocation `bson:",inline"`
	Location     GeoLocation `json:"location" bson:"location" validate:"required"`
	CreatedAt    time.Time   `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt" bson:"updatedAt"`
}

// LocationResponse represents the API response for a location
type LocationResponse struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	BaseLocation `bson:",inline"`
	Latitude     float64 `json:"latitude" bson:"latitude"`
	Longitude    float64 `json:"longitude" bson:"longitude"`
}

// LocationRequest represents the incoming request to create or update a location
type LocationRequest struct {
	BaseLocation `bson:",inline"`
	Latitude     float64 `json:"latitude" bson:"latitude"`
	Longitude    float64 `json:"longitude" bson:"longitude"`
}

type AdInfo struct {
	AdCode          string `json:"adcode" bson:"adcode"`
	CityCode        string `json:"city_code" bson:"city_code"`
	DistrictCode    string `json:"district_code" bson:"district_code"`
	NationCode      string `json:"nation_code" bson:"nation_code"`
	NationalityCode string `json:"nationality_code" bson:"nationality_code"`
}

type AddressComponent struct {
	Nation       string `json:"nation" bson:"nation"`
	Province     string `json:"province" bson:"province"`
	City         string `json:"city" bson:"city"`
	District     string `json:"district" bson:"district"`
	Street       string `json:"street" bson:"street"`
	StreetNumber string `json:"street_number" bson:"street_number"`
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
	Location LocationResponse `json:"location"`
	Distance float64          `json:"distance"` // Distance to the search point in meters
}

// ReverseGeocodeResponse represents the response from Tencent Maps API
type ReverseGeocodeResponse struct {
	Status  int                  `json:"status"`
	Message string               `json:"message"`
	Result  ReverseGeocodeResult `json:"result"`
}

// ReverseGeocodeResult represents the structured result from the geocoding API
type ReverseGeocodeResult struct {
	Location struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"location"`
	Address            string           `json:"address"`
	AddressComponent   AddressComponent `json:"address_component"`
	AdInfo             AdInfo           `json:"ad_info"`
	FormattedAddresses struct {
		Recommend string `json:"recommend"`
		Rough     string `json:"rough"`
	} `json:"formatted_addresses"`
	PoiCount int `json:"poi_count"`
	Pois     []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Address  string `json:"address"`
		Category string `json:"category"`
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
		Distance float64 `json:"_distance"`
	} `json:"pois"`
}
