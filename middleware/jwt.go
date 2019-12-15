package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"server/model/dao"

	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

var config *dao.Config

const authHeader = "X-Auth-Token"

func Setup(cfg *dao.Config) {
	config = cfg
}

func TokenValidation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	if err := jwtValidate(r); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return err
	}

	return nil
}

func jwtValidate(r *http.Request) error {
	tokenString := r.Header.Get(authHeader)

	if tokenString == "" {
		return errors.New("token is empty")
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
		return fmt.Errorf("parse JWT error: %s", err)
	}
	if claims.Exp == 0 || claims.Exp < time.Now().Unix() {
		return errors.New("JWT is expired")
	}

	return nil
}
