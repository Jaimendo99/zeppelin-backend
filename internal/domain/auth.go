package domain

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
type AuthServiceI interface {
	VerifyToken(token string) (*AuthResponse, error)
}
