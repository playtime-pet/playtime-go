package utils

import (
	"fmt"
	"playtime-go/models"
)

// ToGeoJSONPoint converts latitude and longitude into a MongoDB GeoJSON Point
func ToGeoJSONPoint(lat, lon float64) models.GeoLocation {
	return models.GeoLocation{
		Type:        "Point",
		Coordinates: []float64{lon, lat},
	}
}

// FromGeoJSONPoint extracts latitude and longitude from a MongoDB GeoJSON Point
func FromGeoJSONPoint(point models.GeoLocation) (float64, float64, error) {

	coordinates := point.Coordinates
	if len(coordinates) != 2 {
		return 0, 0, fmt.Errorf("invalid GeoJSON Point format")
	}

	lon := coordinates[0]
	lat := coordinates[1]

	return lat, lon, nil
}
