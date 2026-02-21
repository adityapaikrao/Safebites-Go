package model

import "time"

type Favorite struct {
	ID          int       `json:"id"`
	UserID      string    `json:"userId"`
	ProductName string    `json:"productName"`
	Brand       string    `json:"brand,omitempty"`
	SafetyScore *int      `json:"safetyScore,omitempty"`
	Image       string    `json:"image,omitempty"`
	AddedAt     time.Time `json:"addedAt"`
}
