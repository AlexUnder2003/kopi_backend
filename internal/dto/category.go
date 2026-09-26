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

type CategoryResponse struct {
	ID   uuid.UUID `json:"id" db:"id"`
	Name string    `json:"name" db:"name"`
	Icon string    `json:"icon" db:"icon"`
}
