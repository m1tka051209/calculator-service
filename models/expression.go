package models

import "time"

type Expression struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Expression  string     `json:"expression"`
	Status      string     `json:"status"`
	Result      float64    `json:"result,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}