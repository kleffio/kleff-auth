package auth

type SignUpInput struct {
	Tenant    string
	Email     *string
	Username  *string
	Password  string
	AttrsJSON *string
	IP        string
	UserAgent string
}

type SignInInput struct {
	Tenant     string
	Identifier string
	Password   string
	IP         string
	UserAgent  string
}

type TokenOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresInSec int
	TokenType    string
}

type OAuthStartInput struct {
	Provider    string
	Tenant      string
	ClientID    string
	RedirectURI string
	IP          string
	UserAgent   string
}
