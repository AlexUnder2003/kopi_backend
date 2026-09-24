package models

import (
	"github.com/google/uuid"
)

type Category struct {
	ID     uuid.UUID  `db:"id"`
	Name   string     `db:"name"`
	Icon   string     `db:"icon"`
	UserID *uuid.UUID `db:"user_id"`
}
