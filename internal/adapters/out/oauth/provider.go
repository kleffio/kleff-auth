package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	domain "github.com/kleffio/kleff-auth/internal/core/domain/auth"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) BuildAuthURL(ctx context.Context, provider string, cfg *domain.OAuthProviderConfig, state string) (string, error) {
	switch provider {
	case "google":
		return p.buildGoogleAuthURL(cfg, state), nil
	case "github":
		return p.buildGitHubAuthURL(cfg, state), nil
	default:
		return "", errors.New("unsupported provider")
	}
}

func (p *Provider) buildGoogleAuthURL(cfg *domain.OAuthProviderConfig, state string) string {
	params := url.Values{
		"client_id":     {cfg.ClientID},
		"redirect_uri":  {cfg.RedirectURL},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func (p *Provider) buildGitHubAuthURL(cfg *domain.OAuthProviderConfig, state string) string {
	params := url.Values{
		"client_id":    {cfg.ClientID},
		"redirect_uri": {cfg.RedirectURL},
		"scope":        {"user:email"},
		"state":        {state},
	}
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

func (p *Provider) ExchangeCode(ctx context.Context, provider string, cfg *domain.OAuthProviderConfig, code string) (*domain.OAuthIdentity, error) {
	switch provider {
	case "google":
		return p.exchangeGoogleCode(ctx, cfg, code)
	case "github":
		return p.exchangeGitHubCode(ctx, cfg, code)
	default:
		return nil, errors.New("unsupported provider")
	}
}

func (p *Provider) exchangeGoogleCode(ctx context.Context, cfg *domain.OAuthProviderConfig, code string) (*domain.OAuthIdentity, error) {
	// gosec: G101 -- this is a public Google OAuth token endpoint, not a credential
	tokenURL := "https://oauth2.googleapis.com/token"
	data := url.Values{
		"code":          {code},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"redirect_uri":  {cfg.RedirectURL},
		"grant_type":    {"authorization_code"},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &domain.OAuthIdentity{
		Provider:       "google",
		ProviderUserID: userInfo.ID,
		Email:          &userInfo.Email,
		Username:       &userInfo.Name,
		Attrs: map[string]any{
			"picture": userInfo.Picture,
		},
	}, nil
}

func (p *Provider) exchangeGitHubCode(ctx context.Context, cfg *domain.OAuthProviderConfig, code string) (*domain.OAuthIdentity, error) {
	// gosec: G101 -- this is a public Google OAuth token endpoint, not a credential
	tokenURL := "https://github.com/login/oauth/access_token"
	data := url.Values{
		"code":          {code},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"redirect_uri":  {cfg.RedirectURL},
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", tokenURL, nil)
	req.URL.RawQuery = data.Encode()
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("github auth error: %s", tokenResp.Error)
	}

	req, _ = http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()

	var userInfo struct {
		ID        int64   `json:"id"`
		Login     string  `json:"login"`
		Email     *string `json:"email"`
		AvatarURL string  `json:"avatar_url"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &domain.OAuthIdentity{
		Provider:       "github",
		ProviderUserID: fmt.Sprintf("%d", userInfo.ID),
		Email:          userInfo.Email,
		Username:       &userInfo.Login,
		Attrs: map[string]any{
			"avatar_url": userInfo.AvatarURL,
		},
	}, nil
}
