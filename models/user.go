package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	NickName    string             `json:"nickName" bson:"nickName"`
	PhoneNumber string             `json:"phoneNumber" bson:"phoneNumber"`
	AvatarURL   string             `json:"avatarUrl" bson:"avatarUrl"`
	OpenID	  string             `json:"openId" bson:"openId"`
	unionID     string             `json:"unionId" bson:"unionId"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// UserRequest represents the incoming request to create a user
type UserRequest struct {
	NickName    string `json:"nickName"`
	PhoneNumber string `json:"phoneNumber"`
	AvatarURL   string `json:"avatarUrl"`
	OpenID      string `json:"openId"`
	UnionID     string `json:"unionId"`
}
