package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents the 'users' table in your database.
type User struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key"`
	TenantID        uuid.UUID  `gorm:"type:uuid;not null"`
	Email           string     `gorm:"type:varchar(255);not null;unique"`
	Username        string     `gorm:"type:varchar(255);unique"`
	PasswordHash    string     `gorm:"type:varchar(255)"`
	EmailVerifiedAt *time.Time `gorm:"type:timestamptz"`
	Attrs           string     `gorm:"type:jsonb"` // Assuming it's a JSON string
	CreatedAt       time.Time  `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time  `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}

type Client struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null"`
	Name         string    `gorm:"type:varchar(255);not null"`
	RedirectUris string    `gorm:"type:jsonb"` // Assuming it's a JSON string
	Type         string    `gorm:"type:varchar(50);not null"`
	CreatedAt    time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}
type Tenant struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key"`
	Slug           string    `gorm:"type:varchar(255);unique;not null"`
	Name           string    `gorm:"type:varchar(255);not null"`
	Branding       string    `gorm:"type:jsonb"` // Assuming it's a JSON string
	UserAttrSchema string    `gorm:"type:jsonb"` // Assuming it's a JSON string
	CreatedAt      time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}

type Session struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null"`
	ClientID    uuid.UUID  `gorm:"type:uuid;not null"`
	RefreshHash string     `gorm:"type:varchar(255);unique;not null"`
	FamilyID    *uuid.UUID `gorm:"type:uuid"`
	UserAgent   string     `gorm:"type:text"`
	IP          string     `gorm:"type:varchar(100)"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	ExpiresAt   time.Time  `gorm:"type:timestamptz"`
	RevokedAt   *time.Time `gorm:"type:timestamptz"`
}
type Credential struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID     uuid.UUID `gorm:"type:uuid;not null"`
	Kind       string    `gorm:"type:varchar(50);not null"` // e.g., 'password', 'webauthn', 'totp'
	SecretHash string    `gorm:"type:varchar(255)"`
	Meta       string    `gorm:"type:jsonb"` // Assuming it's a JSON string
	CreatedAt  time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}
type SocialIdentity struct {
	UserID          uuid.UUID `gorm:"type:uuid;not null"`
	Provider        string    `gorm:"type:varchar(50);not null"`
	ProviderUID     string    `gorm:"type:varchar(255);not null"`
	EmailAtProvider string    `gorm:"type:varchar(255)"`
	RawProfile      string    `gorm:"type:jsonb"` // Assuming it's a JSON string
}
type SigningKey struct {
	Kid        string    `gorm:"type:varchar(255);primary_key"`
	Alg        string    `gorm:"type:varchar(50);not null"`
	PublicKey  string    `gorm:"type:text;not null"`
	PrivateRef string    `gorm:"type:text"` // Reference to the private key
	CreatedAt  time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}
type AuditLog struct {
	ID        int64      `gorm:"type:bigserial;primary_key"`
	TenantID  uuid.UUID  `gorm:"type:uuid;not null"`
	UserID    *uuid.UUID `gorm:"type:uuid"`                 // Can be null if not available
	Actor     string     `gorm:"type:varchar(50);not null"` // e.g., 'user', 'system', 'admin'
	Event     string     `gorm:"type:varchar(100);not null"`
	IP        string     `gorm:"type:varchar(100)"`
	UserAgent string     `gorm:"type:text"`
	At        time.Time  `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	Details   string     `gorm:"type:jsonb"` // Assuming it's a JSON string
}
type EmailVerificationToken struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null"`
	TokenHash  string     `gorm:"type:varchar(255);unique;not null"`
	ExpiresAt  time.Time  `gorm:"type:timestamptz;not null"`
	ConsumedAt *time.Time `gorm:"type:timestamptz"`
}
