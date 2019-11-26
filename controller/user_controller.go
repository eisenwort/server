package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/core/ewc"
	"server/model/dao"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

// UserCtrl - controller fot user
type UserCtrl struct {
	config  *dao.Config
	service *ewc.DbUserService
}

// NewUserCtrl - create user controller
func NewUserCtrl(cfg *dao.Config) *UserCtrl {
	ctrl := new(UserCtrl)
	ctrl.config = cfg
	ctrl.service = ewc.NewDbUserService(cfg.Driver, cfg.ConnectionString)

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
	// if err := json.NewEncoder(w).Encode(user); err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// }
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dao.JwtClaims{
		Exp: time.Now().Add(srv.tokenLifeTime).Unix(),
		Id:  user.ID,
	})
	tokenString, err := token.SignedString([]byte(srv.vkSecret))

	if err != nil {
		log.Println("create token error", err.Error())
		return ""
	}

	tokenJson := fmt.Sprintf(`{ "token": "%s" }`, token)
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
	// if err := json.NewEncoder(w).Encode(user); err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// }
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dao.JwtClaims{
		Exp: time.Now().Add(srv.tokenLifeTime).Unix(),
		Id:  user.ID,
	})
	tokenString, err := token.SignedString([]byte(srv.vkSecret))

	if err != nil {
		log.Println("create token error", err.Error())
		return ""
	}

	tokenJson := fmt.Sprintf(`{ "token": "%s" }`, token)
	w.Write([]byte(tokenJson))
}

func (ctrl *UserCtrl) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := new(ewc.User)

	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if id, err := strconv.ParseInt(ps.ByName("id"), 10, 64); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return

		if id != user.ID {
			w.WriteHeader(http.StatusForbidden)
			return
		}
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
	if id, err := strconv.ParseInt(ps.ByName("id"), 10, 64); err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
	user := ctrl.service.GetByLogin(ps.ByName("login"))

	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
