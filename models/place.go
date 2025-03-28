package models

import "time"

type Review struct {
	PlaceID    string    `json:"place_id"`
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	Content    string    `json:"content"`
	RatingStar int       `json:"rating_star"`
	Date       time.Time `json:"date"`
}
