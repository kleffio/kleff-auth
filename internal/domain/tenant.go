package domain

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID   uuid.UUID
	Slug string
	Name string

	Branding       map[string]any
	UserAttrSchema map[string]any

	CreatedAt time.Time
	UpdatedAt time.Time
}
