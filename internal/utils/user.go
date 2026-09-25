package utils

import (
	"errors"

	"github.com/google/uuid"

	"github.com/labstack/echo/v5"
)

func GetUserIDFromContext(c *echo.Context) (uuid.UUID, error) {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	return userID, nil
}
