package controller

import (
	"encoding/json"
	"net/http"
	"server/core/ewc"
	"server/model/dao"
	"strconv"

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

	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
	userID, err := strconv.ParseInt(r.Header.Get(core.IdHeader), 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if msg.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !ctrl.service.Delete(msg) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
