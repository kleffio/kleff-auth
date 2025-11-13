package auth

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	UserID     uuid.UUID  `db:"user_id" json:"userId"`
	ClientID   *uuid.UUID `db:"client_id" json:"clientId,omitempty"`
	FamilyID   uuid.UUID  `db:"family_id" json:"familyId"`
	ParentID   *uuid.UUID `db:"parent_id" json:"parentId,omitempty"`
	ReplacedBy *uuid.UUID `db:"replaced_by" json:"replacedBy,omitempty"`

	RefreshHash []byte `db:"refresh_hash" json:"-"`
	UserAgent   string `db:"user_agent" json:"userAgent"`
	IP          net.IP `db:"ip" json:"ip"`

	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	LastUsedAt time.Time  `db:"last_used_at" json:"lastUsedAt"`
	ExpiresAt  time.Time  `db:"expires_at" json:"expiresAt"`
	RevokedAt  *time.Time `db:"revoked_at" json:"revokedAt,omitempty"`
	Reason     *string    `db:"reason" json:"reason,omitempty"`
}
