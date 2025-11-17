package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Email        *string
	Username     *string
	PasswordHash string
	Attrs        map[string]any

	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
