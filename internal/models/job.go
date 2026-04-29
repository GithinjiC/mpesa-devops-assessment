package models

import "time"

type Job struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Company     *string   `json:"company,omitempty"`
	Description *string   `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
