package domain

import (
	"time"

	"github.com/google/uuid"
)

type Flag struct {
	ID        uuid.UUID
	Key       string
	Config    Config
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
