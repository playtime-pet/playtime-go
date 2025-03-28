package models

import "time"

type Review struct {
	PlaceID    string    `json:"place_id"`
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	UserAvatar string    `json:"user_avatar"`
	Content    string    `json:"content"`
	Rating     int       `json:"rating"`
	Date       time.Time `json:"date"`
}
