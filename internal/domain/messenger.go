package domain

import "time"

// UserMessage represents a normalized message from any messenger source
type UserMessage struct {
	UserID    string                 `json:"user_id"`
	Content   string                 `json:"content"`
	Source    string                 `json:"source"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// MessageResponse represents a standard response to be sent back to the user
type MessageResponse struct {
	Text string      `json:"text"`
	Data interface{} `json:"data,omitempty"`
}
