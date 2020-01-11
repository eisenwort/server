package dao

import (
	"server/core/ewc"

	"github.com/dgrijalva/jwt-go"
)

// Config - app config
type Config struct {
	ServiceAddress   string `json:"service_address"`
	Driver           string `json:"driver"`
	ConnectionString string `json:"connection_string"`
	JwtSign          string `json:"jwt_sign"`
	PageLimit        int    `json:"page_limit"`
}

type ApiError struct {
	Error string `json:"error,omitempty"`
}

type AuthData struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type JwtClaims struct {
	*jwt.MapClaims
	Id  int64
	Exp int64
}

func (JwtClaims) Valid() error {
	return nil
}

type ChatData struct {
	ewc.Chat
	UnreadCount int `json:"unread_count"`
}
