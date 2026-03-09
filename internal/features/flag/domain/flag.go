package domain

import "time"

type Flag struct {
	Key       string
	Config    Config
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
