package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"server/core/ewc"
	"server/model/dao"

	"github.com/gorilla/mux"
)

type MessageCtrl struct {
	config      *dao.Config
	service     *ewc.DbMessageService
	chatService *ewc.DbChatService
}

func NewMessageCtrl(cfg *dao.Config) *MessageCtrl {
	ctrl := new(MessageCtrl)
	ctrl.config = cfg
	ctrl.service = ewc.NewDbMessageService()
	ctrl.chatService = ewc.NewDbChatService()

	return ctrl
}

func (ctrl MessageCtrl) Create(w http.ResponseWriter, r *http.Request) {
	msg := ewc.Message{}
	claims := getClaims(r)

	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if msg.UserID != claims.Id {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !ctrl.chatService.IsUserInChat(msg.ChatID, claims.Id) {
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

func (ctrl MessageCtrl) Delete(w http.ResponseWriter, r *http.Request) {
	msg := ewc.Message{}
	claims := getClaims(r)
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

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

func (ctrl MessageCtrl) GetByChat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatId, err := strconv.ParseInt(vars["chat_id"], 10, 64)

	if err != nil {
		log.Println("parse id error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	page, err := strconv.Atoi(r.FormValue("page"))

	if err != nil {
		log.Println("parse page error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := getClaims(r)

	if !ctrl.chatService.IsUserInChat(chatId, claims.Id) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	messages := ctrl.service.GetByChat(chatId, page)

	if err := json.NewEncoder(w).Encode(messages); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ctrl MessageCtrl) GetLastId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatId, err := strconv.ParseInt(vars["chat_id"], 10, 64)

	if err != nil {
		log.Println("parse id error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	lastId := ctrl.service.GetLastId(chatId)

	w.Header().Set("X-Last-Id", strconv.FormatInt(lastId, 10))
}
