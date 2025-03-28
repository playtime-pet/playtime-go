package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Pet represents a pet in the system
type Pet struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Gender    string             `json:"gender" bson:"gender"`
	Size      string             `json:"size" bson:"size"`
	Breed     string             `json:"breed" bson:"breed"`
	Avatar    string             `json:"avatar" bson:"avatar"`
	Character string             `json:"character" bson:"character"`
	Age       int                `json:"age" bson:"age"`
	OwnerID   primitive.ObjectID `json:"ownerId,omitempty" bson:"ownerId,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// PetRequest represents the incoming request to create or update a pet
type PetRequest struct {
	Name      string             `json:"name"`
	Gender    string             `json:"gender"`
	Size      string             `json:"size"`
	Breed     string             `json:"breed"`
	Avatar    string             `json:"avatar"`
	Character string             `json:"character"`
	Age       int                `json:"age"`
	OwnerID   primitive.ObjectID `json:"ownerId,omitempty"`
}
