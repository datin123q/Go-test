package controllers

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	ExpireToken  string `json:"expire_token"`
	UserId       int64  `json:"user_id"`
}

type TokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
