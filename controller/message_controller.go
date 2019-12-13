package controller

import (
	"encoding/json"
	"net/http"

	"server/core/ewc"
	"server/model/dao"

	"github.com/julienschmidt/httprouter"
)

type MessageCtrl struct {
	config  *dao.Config
	service *ewc.DbMessageService
}

func NewMessageCtrl(cfg *dao.Config) *MessageCtrl {
	ctrl := new(MessageCtrl)
	ctrl.config = cfg
	ctrl.service = ewc.NewDbMessageService(cfg.Driver, cfg.ConnectionString)

	return ctrl
}

func (ctrl MessageCtrl) Create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	msg := new(ewc.Message)
	claims := getClaims(r)

	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if msg.UserID != claims.Id {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	_, err := ctrl.service.Create(msg)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	/*if err := json.NewEncoder(w).Encode(item); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}*/
	// todo: to socket
}

func (ctrl MessageCtrl) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	msg := new(ewc.Message)
	claims := getClaims(r)

	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if msg.UserID != claims.Id {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !ctrl.service.Delete(msg) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
