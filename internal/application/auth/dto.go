package auth

type SignUpInput struct {
	Tenant    string
	Email     *string
	Username  *string
	Password  string
	AttrsJSON []byte
}

type SignInInput struct {
	Tenant     string
	Identifier string
	Password   string
}

type TokenOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresInSec int
	TokenType    string
}
