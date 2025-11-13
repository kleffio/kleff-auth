package auth

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	domain "github.com/kleffio/kleff-auth/internal/core/domain/auth"
	"github.com/kleffio/kleff-auth/internal/core/port/auth"
)

var (
	ErrUnknownTenant       = errors.New("unknown tenant")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefresh      = errors.New("invalid refresh")
	ErrReuseDetected       = errors.New("refresh reuse detected")
	ErrInvalidClient       = errors.New("invalid oauth client")
	ErrInvalidRedirectURI  = errors.New("invalid redirect uri")
	ErrUnsupportedProvider = errors.New("unsupported oauth provider")
	ErrInvalidState        = errors.New("invalid oauth state")
)

type Service struct {
	Tenants  auth.TenantRepoPort
	Users    auth.UserRepoPort
	Hash     auth.PasswordHasherPort
	Tokens   auth.TokenSignerPort
	Sessions auth.SessionRepoPort
	Refresh  auth.RefreshTokenPort
	Time     auth.TimeProvider

	OAuthProviders auth.OAuthProviderPort
	OAuthUsers     auth.OAuthUserRepoPort
	OAuthClients   auth.OAuthClientRepoPort
	OAuthState     auth.OAuthStateCodecPort

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

func (s *Service) BuildOAuthRedirectURL(ctx context.Context, in OAuthStartInput) (string, error) {
	if s.OAuthProviders == nil || s.OAuthClients == nil || s.OAuthState == nil {
		return "", errors.New("oauth not configured")
	}

	if in.Tenant == "" {
		return "", ErrUnknownTenant
	}

	tenantID, err := s.Tenants.ResolveTenantID(ctx, in.Tenant)
	if err != nil {
		return "", ErrUnknownTenant
	}

	client, err := s.OAuthClients.GetOAuthClient(ctx, tenantID, in.ClientID, in.Provider)
	if err != nil {
		return "", ErrInvalidClient
	}

	if !client.AllowsRedirect(in.RedirectURI) {
		return "", ErrInvalidRedirectURI
	}

	providerCfg, ok := client.Providers[in.Provider]
	if !ok {
		return "", ErrUnsupportedProvider
	}

	statePayload := &domain.OAuthState{
		TenantID:    tenantID,
		TenantSlug:  in.Tenant,
		ClientID:    in.ClientID,
		RedirectURI: in.RedirectURI,
		Nonce:       uuid.NewString(),
		IssuedAt:    s.nowUTC(),
	}

	stateStr, err := s.OAuthState.Encode(statePayload)
	if err != nil {
		return "", err
	}

	return s.OAuthProviders.BuildAuthURL(ctx, in.Provider, &providerCfg, stateStr)
}

func (s *Service) HandleOAuthCallback(
	ctx context.Context,
	provider, code, stateStr, ip, userAgent string,
) (userID string, email, username *string, tok TokenOutput, err error) {
	if s.OAuthProviders == nil || s.OAuthUsers == nil || s.OAuthState == nil || s.OAuthClients == nil {
		err = errors.New("oauth not configured")
		return
	}

	state, err := s.OAuthState.Decode(stateStr)
	if err != nil {
		err = ErrInvalidState
		return
	}

	client, err := s.OAuthClients.GetOAuthClient(ctx, state.TenantID, state.ClientID, provider)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrInvalidClient
		}
		return
	}

	providerCfg, ok := client.Providers[provider]
	if !ok {
		err = ErrUnsupportedProvider
		return
	}

	identity, err := s.OAuthProviders.ExchangeCode(ctx, provider, &providerCfg, code)
	if err != nil {
		return
	}

	if identity != nil && identity.TenantSlug == "" {
		identity.TenantSlug = state.TenantSlug
	}

	var (
		tenantID = state.TenantID
		uid      string
		em, un   *string
	)

	if tID, uID, emFound, unFound, e := s.OAuthUsers.GetUserByOAuth(
		ctx,
		identity.Provider,
		identity.ProviderUserID,
	); e == nil && uID != "" {
		tenantID = tID
		uid = uID
		em = emFound
		un = unFound
	} else {
		tID, uID, e := s.OAuthUsers.CreateUserFromOAuth(ctx, identity)
		if e != nil {
			err = e
			return
		}
		tenantID = tID
		uid = uID
		em = identity.Email
		un = identity.Username
	}

	jti := uuid.NewString()
	access, err := s.Tokens.IssueAccess(uid, tenantID, "openid profile", jti, s.AccessTTL)
	if err != nil {
		return
	}

	sid := uuid.New()
	raw, rhash, err := s.Refresh.Generate()
	if err != nil {
		return
	}

	now := s.nowUTC()
	sess := &domain.Session{
		ID:          sid,
		UserID:      uuid.MustParse(uid),
		FamilyID:    uuid.New(),
		RefreshHash: rhash,
		UserAgent:   userAgent,
		IP:          net.ParseIP(ip),
		CreatedAt:   now,
		LastUsedAt:  now,
		ExpiresAt:   now.Add(s.RefreshTTL),
	}
	if err = s.Sessions.Create(ctx, sess); err != nil {
		return
	}

	refresh := s.Refresh.Encode(sid, raw)

	tok = TokenOutput{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresInSec: int(s.AccessTTL.Seconds()),
		TokenType:    "Bearer",
	}

	userID = uid
	email = em
	username = un
	return
}

func (s *Service) nowUTC() time.Time {
	if s.Time != nil {
		return s.Time.Now().UTC()
	}
	return time.Now().UTC()
}
