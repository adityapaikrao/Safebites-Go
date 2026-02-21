package model

import "time"

type Scan struct {
	ID          string                   `json:"id"`
	UserID      string                   `json:"userId"`
	ProductName string                   `json:"productName"`
	Brand       string                   `json:"brand,omitempty"`
	Image       string                   `json:"image,omitempty"`
	SafetyScore int                      `json:"safetyScore"`
	IsSafe      bool                     `json:"isSafe"`
	Ingredients []map[string]interface{} `json:"ingredients"`
	Timestamp   time.Time                `json:"timestamp"`
}
