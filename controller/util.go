package controller

import (
	"log"
	"net/http"
	"strings"

	"server/model/dao"

	"github.com/dgrijalva/jwt-go"
)

var Config *dao.Config

func getClaims(r *http.Request) dao.JwtClaims {
	claims := dao.JwtClaims{}
	token := r.Header.Get("X-Auth-Token")

	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(Config.JwtSign), nil
	})

	if err != nil {
		log.Println("parse JWT error:", err)
	}

	return claims
}

func getInclude(include string) []string {
	return strings.Split(include, ",")
}
