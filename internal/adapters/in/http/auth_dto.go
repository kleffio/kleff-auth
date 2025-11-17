package http

type signUpRequestDTO struct {
	Tenant    string  `json:"tenant"`
	Email     *string `json:"email,omitempty"`
	Username  *string `json:"username,omitempty"`
	Password  string  `json:"password"`
	AttrsJSON *string `json:"attrs_json,omitempty"`
}

type signInRequestDTO struct {
	Tenant     string `json:"tenant"`
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type tokenResponseDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresInSec int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}
