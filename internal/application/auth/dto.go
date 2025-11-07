package auth

type SignUpInput struct {
	Tenant    string  `json:"tenant"`
	Email     *string `json:"email,omitempty"`
	Username  *string `json:"username,omitempty"`
	Password  string  `json:"password"`
	AttrsJSON *string `json:"attrs_json,omitempty"`
	IP        string  `json:"-"`
	UserAgent string  `json:"-"`
}

type SignInInput struct {
	Tenant     string `json:"tenant"`
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
	IP         string `json:"-"`
	UserAgent  string `json:"-"`
}

type TokenOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresInSec int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}
