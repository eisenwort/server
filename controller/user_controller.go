package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"server/core/ewc"
	"server/model/dao"

	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

// UserCtrl - controller fot user
type UserCtrl struct {
	config        *dao.Config
	service       *ewc.DbUserService
	tokenLifeTime time.Duration
}

// NewUserCtrl - create user controller
func NewUserCtrl(cfg *dao.Config) *UserCtrl {
	ctrl := new(UserCtrl)
	ctrl.config = cfg
	ctrl.service = ewc.NewDbUserService(cfg.Driver, cfg.ConnectionString)
	ctrl.tokenLifeTime = 1 * time.Hour

	return ctrl
}

// Login - auth user
func (ctrl *UserCtrl) Login(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data := make(map[string]string)
	var login, password string
	ok := false

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if login, ok = data["login"]; !ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if password, ok = data["password"]; !ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	user := ctrl.service.Login(login, password)

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dao.JwtClaims{
		Exp: time.Now().Add(ctrl.tokenLifeTime).Unix(),
		Id:  user.ID,
	})
	tokenString, err := token.SignedString([]byte(ctrl.config.JwtSign))

	if err != nil {
		log.Println("create token error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenJson := fmt.Sprintf(`{ "token": "%s" }`, tokenString)
	w.Write([]byte(tokenJson))
}

// Login - auth user
func (ctrl *UserCtrl) Registration(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data := make(map[string]string)
	var login, password string
	ok := false

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if login, ok = data["login"]; !ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if password, ok = data["password"]; !ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	user, err := ctrl.service.Create(login, password)

	if user == nil {
		errData, _ := json.Marshal(&dao.ApiError{Error: err.Error()})
		w.WriteHeader(http.StatusConflict)
		w.Write(errData)
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dao.JwtClaims{
		Exp: time.Now().Add(ctrl.tokenLifeTime).Unix(),
		Id:  user.ID,
	})
	tokenString, err := token.SignedString([]byte(ctrl.config.JwtSign))

	if err != nil {
		log.Println("create token error", err.Error())
		return
	}

	tokenJson := fmt.Sprintf(`{ "token": "%s" }`, tokenString)
	w.Write([]byte(tokenJson))
}

func (ctrl *UserCtrl) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := new(ewc.User)
	claims := getClaims(r)

	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if id != user.ID || id != claims.Id {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	user = ctrl.service.Update(user)

	if user == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (ctrl *UserCtrl) Get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := getClaims(r)

	if id != claims.Id {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	user := ctrl.service.Get(id)

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (ctrl *UserCtrl) GetByLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	claims := getClaims(r)
	user := ctrl.service.GetByLogin(ps.ByName("login"))

	if user.ID != claims.Id {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
