package controller

import (
	"encoding/json"
	"net/http"
	"server/core/ewc"
	"server/model/dao"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type ChatCtrl struct {
	config  *dao.Config
	service *ewc.DbChatService
}

func NewChatCtrl(cfg *dao.Config) *ChatCtrl {
	ctrl := new(ChatCtrl)
	ctrl.config = cfg
	ctrl.service = ewc.NewDbChatService(cfg.Driver, cfg.ConnectionString)

	return ctrl
}

func (ctrl *ChatCtrl) GetList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID, err := strconv.ParseInt(r.Header.Get("X-Auth-Id"), 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	chats, err := ctrl.service.GetForUser(userID)

	// todo: unread messages

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(&chats); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ctrl *ChatCtrl) Get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if status := ctrl.checkRights(w, r, ps); status != 0 {
		w.WriteHeader(status)
		return
	}

	id, _ := strconv.ParseInt(ps.ByName("id"), 10, 64)
	chat, _ := ctrl.service.Get(id, true)

	if err := json.NewEncoder(w).Encode(chat); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ctrl *ChatCtrl) Create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID, err := strconv.ParseInt(r.Header.Get("X-Auth-Id"), 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	chat := &ewc.Chat{}
	isExist := false

	if err := json.NewDecoder(r.Body).Decode(chat); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, user := range chat.Users {
		if user.ID == userID {
			isExist = true
		}
	}
	if !isExist {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	item, err := ctrl.service.Create(chat)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(chat); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ctrl *ChatCtrl) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if status := ctrl.checkRights(w, r, ps); status != 0 {
		w.WriteHeader(status)
		return
	}

	id, _ := strconv.ParseInt(ps.ByName("id"), 10, 64)
	chat, _ := ctrl.service.Get(id, true)

	if !chat.Personal || chat.OwnerID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ctrl.service.Delete(chat)
}

func (ctrl *ChatCtrl) Exit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if status := ctrl.checkRights(w, r, ps); status != 0 {
		w.WriteHeader(status)
		return
	}

	id, _ := strconv.ParseInt(ps.ByName("id"), 10, 64)
	chat, _ := ctrl.service.Get(id, true)
	ctrl.service.Exit(chat)
}

func (ctrl *ChatCtrl) Clean(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if status := ctrl.checkRights(w, r, ps); status != 0 {
		w.WriteHeader(status)
		return
	}

	id, _ := strconv.ParseInt(ps.ByName("id"), 10, 64)
	chat, _ := ctrl.service.Get(id, true)

	if !ctrl.service.Clean(chat) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ctrl *ChatCtrl) checkRights(w http.ResponseWriter, r *http.Request, ps httprouter.Params) int {
	userID, err := strconv.ParseInt(r.Header.Get("X-Auth-Id"), 10, 64)

	if err != nil {
		return http.StatusForbidden
	}

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil {
		return http.StatusBadRequest
	}

	chat, err := ctrl.service.Get(id, true)
	isExist := false

	if err != nil {
		return http.StatusInternalServerError
	}
	for _, user := range chat.Users {
		if user.ID == userID {
			isExist = true
		}
	}
	if !isExist {
		return http.StatusForbidden
	}

	return 0
}
