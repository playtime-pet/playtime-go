package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Review struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	PlaceID    string             `json:"place_id" bson:"placeId"`
	UserID     string             `json:"user_id" bson:"userId"`
	UserName   string             `json:"user_name" bson:"userName"`
	UserAvatar string             `json:"user_avatar" bson:"userAvatar"`
	Content    string             `json:"content" bson:"content"`
	Rating     int                `json:"rating" bson:"rating"`
	Date       time.Time          `json:"date" bson:"date"`
}
