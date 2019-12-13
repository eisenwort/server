package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"server/model/dao"

	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

var config *dao.Config

const authHeader = "Authorization"
const bearer = "Bearer "

func Setup(cfg *dao.Config) {
	config = cfg
}

func TokenValidation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	token := r.Header.Get(authHeader)
	result := false

	if len(token) != 0 {
		result = jwtValidate(r)
	}
	if !result {
		return errors.New("")
	}

	return nil
}

func jwtValidate(r *http.Request) bool {
	tokenString := getJwtString(r)

	if tokenString == "" {
		return false
	}

	claims := dao.JwtClaims{}
	_, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(config.JwtSign), nil
		},
	)

	if err != nil {
		log.Println("JWT error:", err)
		return false
	}
	if claims.Exp == 0 || claims.Exp < time.Now().Unix() {
		return false
	}

	return true
}

func getJwtString(r *http.Request) string {
	headerValue := r.Header.Get(authHeader)

	if !strings.HasPrefix(headerValue, bearer) {
		return ""
	}

	return strings.TrimPrefix(headerValue, bearer)
}
