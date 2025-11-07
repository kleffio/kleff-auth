package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ClientID    *uuid.UUID
	RefreshHash string
	FamilyID    *uuid.UUID
	UserAgent   string
	IP          string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	RevokedAt   *time.Time
}
