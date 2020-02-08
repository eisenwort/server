package controller

import (
	"encoding/json"
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
	ctrl.service = ewc.NewDbUserService()
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

	jsonData, _ := json.Marshal(ctrl.createAuthData(user.ID))
	w.Write(jsonData)
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
	if password, ok = data["reset_password"]; !ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	user, err := ctrl.service.Create(login, password)

	if err != nil {
		log.Println("create user error:", err)
		errData, _ := json.Marshal(&dao.ApiError{Error: err.Error()})
		w.WriteHeader(http.StatusConflict)
		w.Write(errData)
		return
	}

	jsonData, _ := json.Marshal(ctrl.createAuthData(user.ID))
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func (ctrl *UserCtrl) RefreshToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	claims := getClaims(r)
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if claims.Id == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if claims.Id != id {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	jsonData, _ := json.Marshal(ctrl.createAuthData(id))
	w.Write(jsonData)
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

func (ctrl *UserCtrl) createToken(id int64, duration time.Duration) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dao.JwtClaims{
		Exp: time.Now().Add(duration).Unix(),
		Id:  id,
	})
	tokenString, _ := token.SignedString([]byte(ctrl.config.JwtSign))

	return tokenString
}

func (ctrl *UserCtrl) createAuthData(id int64) *dao.AuthData {
	return &dao.AuthData{
		Token:        ctrl.createToken(id, ctrl.tokenLifeTime),
		RefreshToken: ctrl.createToken(id, 336*time.Hour),
	}
}
