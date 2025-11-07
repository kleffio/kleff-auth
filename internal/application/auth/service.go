package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUnknownTenant      = errors.New("unknown tenant")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	Tenants TenantRepoPort
	Users   UserRepoPort
	Hash    PasswordHasherPort
	Tokens  TokenSignerPort

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

	hash, err := s.Hash.Hash(in.Password)
	if err != nil {
		return "", tok, err
	}

	userID, err = s.Users.CreateUser(ctx, tenantID, in.Email, in.Username, hash, in.AttrsJSON)
	if err != nil {
		return "", tok, err
	}

	jti := uuid.NewString()
	access, err := s.Tokens.IssueAccess(userID, tenantID, "openid profile", jti, s.AccessTTL)
	if err != nil {
		return "", tok, err
	}

	refreshRaw, _, err := s.Tokens.NewRefresh()
	if err != nil {
		return "", tok, err
	}

	return userID, TokenOutput{AccessToken: access, RefreshToken: refreshRaw, ExpiresInSec: int(s.AccessTTL.Seconds()), TokenType: "Bearer"}, nil
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
	acc, err := s.Tokens.IssueAccess(uid, tenantID, "openid profile", jti, s.AccessTTL)
	if err != nil {
		return "", nil, nil, tok, err
	}

	ref, _, err := s.Tokens.NewRefresh()
	if err != nil {
		return "", nil, nil, tok, err
	}

	return uid, em, un, TokenOutput{AccessToken: acc, RefreshToken: ref, ExpiresInSec: int(s.AccessTTL.Seconds()), TokenType: "Bearer"}, nil
}
