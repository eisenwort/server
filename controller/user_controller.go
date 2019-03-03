package controller

import (
	"encoding/json"
	"net/http"
	"server/core/ewc"
	"server/model/dao"
	"strconv"

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
	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
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

	user, _ := ctrl.service.Create(login, password)

	if user == nil {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
}
