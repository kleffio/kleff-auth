package auth

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	domain "github.com/kleffio/kleff-auth/internal/core/domain/auth"
	auth2 "github.com/kleffio/kleff-auth/internal/core/port/auth"
)

var (
	ErrUnknownTenant      = errors.New("unknown tenant")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRefresh     = errors.New("invalid refresh")
	ErrReuseDetected      = errors.New("refresh reuse detected")
)

type Service struct {
	Tenants  auth2.TenantRepoPort
	Users    auth2.UserRepoPort
	Hash     auth2.PasswordHasherPort
	Tokens   auth2.TokenSignerPort
	Sessions auth2.SessionRepoPort
	Refresh  auth2.RefreshTokenPort
	Time     auth2.TimeProvider

	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func (s *Service) JWKS() []byte { return s.Tokens.JWKS() }

func (s *Service) SignUp(ctx context.Context, in SignUpInput) (userID string, tok TokenOutput, err error) {
	if in.Tenant == "" || in.Password == "" || (in.Email == nil && in.Username == nil) {
		return "", tok, errors.New("missing fields")
	}

	tenantID, err := s.Tenants.ResolveTenantID(ctx, in.Tenant)
	if err != nil {
		return "", tok, ErrUnknownTenant
	}

	passHash, err := s.Hash.Hash(in.Password)
	if err != nil {
		return "", tok, err
	}

	userID, err = s.Users.CreateUser(ctx, tenantID, in.Email, in.Username, passHash, in.AttrsJSON)
	if err != nil {
		return "", tok, err
	}

	jti := uuid.NewString()
	access, err := s.Tokens.IssueAccess(userID, tenantID, "openid profile", jti, s.AccessTTL)
	if err != nil {
		return "", tok, err
	}

	sid := uuid.New()
	raw, rhash, err := s.Refresh.Generate()
	if err != nil {
		return "", tok, err
	}

	now := s.nowUTC()
	sess := &domain.Session{
		ID:          sid,
		UserID:      uuid.MustParse(userID),
		FamilyID:    uuid.New(),
		RefreshHash: rhash,
		UserAgent:   in.UserAgent,
		IP:          net.ParseIP(in.IP),
		CreatedAt:   now,
		LastUsedAt:  now,
		ExpiresAt:   now.Add(s.RefreshTTL),
	}
	if err := s.Sessions.Create(ctx, sess); err != nil {
		return "", tok, err
	}

	refresh := s.Refresh.Encode(sid, raw)

	return userID, TokenOutput{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresInSec: int(s.AccessTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) SignIn(ctx context.Context, in SignInInput) (userID string, email, username *string, tok TokenOutput, err error) {
	if in.Tenant == "" || in.Identifier == "" || in.Password == "" {
		return "", nil, nil, tok, errors.New("missing fields")
	}

	tenantID, err := s.Tenants.ResolveTenantID(ctx, in.Tenant)
	if err != nil {
		return "", nil, nil, tok, ErrUnknownTenant
	}

	uid, passHash, em, un, err := s.Users.GetUserByIdentifier(ctx, tenantID, in.Identifier)
	if err != nil {
		return "", nil, nil, tok, ErrInvalidCredentials
	}

	ok, _ := s.Hash.Verify(in.Password, passHash)
	if !ok {
		return "", nil, nil, tok, ErrInvalidCredentials
	}

	if s.Hash.NeedsRehash(passHash) {
		if newH, err := s.Hash.Hash(in.Password); err == nil {
			_ = s.Users.UpdatePasswordHash(ctx, uid, newH)
		}
	}

	jti := uuid.NewString()
	access, err := s.Tokens.IssueAccess(uid, tenantID, "openid profile", jti, s.AccessTTL)
	if err != nil {
		return "", nil, nil, tok, err
	}

	sid := uuid.New()
	raw, rhash, err := s.Refresh.Generate()
	if err != nil {
		return "", nil, nil, tok, err
	}

	now := s.nowUTC()
	sess := &domain.Session{
		ID:          sid,
		UserID:      uuid.MustParse(uid),
		FamilyID:    uuid.New(),
		RefreshHash: rhash,
		UserAgent:   in.UserAgent,
		IP:          net.ParseIP(in.IP),
		CreatedAt:   now,
		LastUsedAt:  now,
		ExpiresAt:   now.Add(s.RefreshTTL),
	}
	if err := s.Sessions.Create(ctx, sess); err != nil {
		return "", nil, nil, tok, err
	}

	refresh := s.Refresh.Encode(sid, raw)

	return uid, em, un, TokenOutput{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresInSec: int(s.AccessTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) RefreshTokens(ctx context.Context, refreshToken string, ua string, ip string, tenantID string) (tok TokenOutput, err error) {
	if refreshToken == "" {
		return tok, ErrInvalidRefresh
	}

	sid, secret, err := s.Refresh.Parse(refreshToken)
	if err != nil {
		return tok, ErrInvalidRefresh
	}

	old, err := s.Sessions.FindByID(ctx, sid)
	if err != nil || old == nil || old.RevokedAt != nil || time.Now().After(old.ExpiresAt) {
		return tok, ErrInvalidRefresh
	}

	replaced, err := s.Sessions.ReplacedAlready(ctx, old.ID)
	if err != nil {
		return tok, err
	}
	if replaced {
		_ = s.Sessions.RevokeFamily(ctx, old.FamilyID, "reuse_detected")
		return tok, ErrReuseDetected
	}

	if err := s.Refresh.Verify(old.RefreshHash, secret); err != nil {
		return tok, ErrInvalidRefresh
	}

	newSecret, newHash, err := s.Refresh.Generate()
	if err != nil {
		return tok, err
	}

	now := s.nowUTC()
	newSess := &domain.Session{
		ID:          uuid.New(),
		UserID:      old.UserID,
		FamilyID:    old.FamilyID,
		ParentID:    &old.ID,
		RefreshHash: newHash,
		UserAgent:   ua,
		IP:          net.ParseIP(ip),
		CreatedAt:   now,
		LastUsedAt:  now,
		ExpiresAt:   now.Add(s.RefreshTTL),
	}
	if err := s.Sessions.Create(ctx, newSess); err != nil {
		return tok, err
	}
	if err := s.Sessions.MarkReplaced(ctx, old.ID, newSess.ID); err != nil {
		return tok, err
	}

	if tenantID == "" {
		if tenantID, err = s.Users.GetTenantIDByUser(ctx, old.UserID.String()); err != nil || tenantID == "" {
			return tok, ErrInvalidRefresh
		}
	}

	jti := uuid.NewString()
	access, err := s.Tokens.IssueAccess(old.UserID.String(), tenantID, "openid profile", jti, s.AccessTTL)
	if err != nil {
		return tok, err
	}

	return TokenOutput{
		AccessToken:  access,
		RefreshToken: s.Refresh.Encode(newSess.ID, newSecret),
		ExpiresInSec: int(s.AccessTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return ErrInvalidRefresh
	}
	sid, _, err := s.Refresh.Parse(refreshToken)
	if err != nil {
		return ErrInvalidRefresh
	}
	return s.Sessions.Revoke(ctx, sid, "user_logout")
}

func (s *Service) LogoutAll(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return ErrInvalidRefresh
	}
	sid, _, err := s.Refresh.Parse(refreshToken)
	if err != nil {
		return ErrInvalidRefresh
	}
	old, err := s.Sessions.FindByID(ctx, sid)
	if err != nil || old == nil {
		return ErrInvalidRefresh
	}
	return s.Sessions.RevokeFamily(ctx, old.FamilyID, "user_logout_all")
}

func (s *Service) Me(ctx context.Context, tenantID, userID string) (email *string, username *string, err error) {
	return s.Users.GetUserByID(ctx, tenantID, userID)
}

func (s *Service) nowUTC() time.Time {
	if s.Time != nil {
		return s.Time.Now().UTC()
	}
	return time.Now().UTC()
}
