package dto

import (
	"errors"

	"github.com/google/uuid"
)

type CategoryPost struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

func (r *CategoryPost) Validate() error {
	if r.Name == "" {
		return errors.New("name")
	}
	return nil
}

type CategoryUpdate struct {
	Name string `json:"name,omitempty"`
	Icon string `json:"icon,omitempty"`
}

type CategoryResponseList struct {
	Categories []CategoryResponse `json:"categories"`
	TotalCount int                `json:"total_count"`
}

type CategoryResponse struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	Icon     string    `json:"icon" db:"icon"`
	IsSystem bool      `json:"is_system" db:"is_system"`
	UserID   uuid.UUID `json:"-" db:"user_id"`
}
