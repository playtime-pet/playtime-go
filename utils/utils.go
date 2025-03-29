package utils

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	return primitive.NewObjectID().Hex()[:8]
}

func ExtractUrlParam(url string, prefix string) []string {
	urlPath := strings.TrimPrefix(url, prefix)
	urlPath = strings.TrimPrefix(urlPath, "/")
	// Split the URL path into parts
	urlParts := strings.Split(urlPath, "/")
	partList := make([]string, 0)
	for _, part := range urlParts {
		if part != "" {
			partList = append(partList, part)
		}
	}

	return partList
}
