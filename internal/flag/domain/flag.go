package domain

import (
	"time"

	"github.com/google/uuid"
)

type Flag struct {
	ID          uuid.UUID
	Key         string
	Enabled     bool
	Description string
	Config      Config
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
