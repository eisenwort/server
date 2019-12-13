package controller

import (
	"net/http"
	"strings"

	"server/model/dao"

	"github.com/dgrijalva/jwt-go"
)

var Config *dao.Config

func getClaims(r *http.Request) dao.JwtClaims {
	claims := dao.JwtClaims{}
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))

	_, _ = jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(Config.JwtSign), nil
	})

	return claims
}
