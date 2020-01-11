package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
	ctrl.service = ewc.NewDbMessageService()

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
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil {
		log.Println("parse id error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if msg.ID != id {
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
