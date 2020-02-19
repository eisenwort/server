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

type ChatCtrl struct {
	config      *dao.Config
	service     *ewc.DbChatService
	userService *ewc.DbUserService
}

func NewChatCtrl(cfg *dao.Config) *ChatCtrl {
	ctrl := new(ChatCtrl)
	ctrl.config = cfg
	ctrl.service = ewc.NewDbChatService()
	ctrl.userService = ewc.NewDbUserService()

	return ctrl
}

func (ctrl *ChatCtrl) GetList(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	chats, err := ctrl.service.GetForUser(claims.Id)

	// TODO: unread messages

	if err != nil {
		log.Println("get chat list for user error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(&chats); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ctrl *ChatCtrl) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if err != nil {
		log.Println("parse id error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if status := ctrl.checkRights(id, r); status != 0 {
		w.WriteHeader(status)
		return
	}

	includes := getInclude(vars["include"])
	chat, err := ctrl.service.Get(id, includes)

	if err := json.NewEncoder(w).Encode(chat); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ctrl *ChatCtrl) Create(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)

	if user := ctrl.userService.Get(claims.Id); user.Reseted {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chat := ewc.Chat{}
	isExist := false

	if err := json.NewDecoder(r.Body).Decode(&chat); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, user := range chat.Users {
		if user.ID == claims.Id {
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

	if err := json.NewEncoder(w).Encode(item); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ctrl *ChatCtrl) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if err != nil {
		log.Println("parse id for delete error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if status := ctrl.checkRights(id, r); status != 0 {
		w.WriteHeader(status)
		return
	}

	claims := getClaims(r)
	chat, err := ctrl.service.Get(id, []string{})

	if err != nil {
		log.Println("get chat for delete error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !chat.Personal || chat.OwnerID != claims.Id {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ctrl.service.Delete(&chat)
}

func (ctrl *ChatCtrl) Exit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if err != nil {
		log.Println("parse id for exit error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if status := ctrl.checkRights(id, r); status != 0 {
		w.WriteHeader(status)
		return
	}

	chat, err := ctrl.service.Get(id, []string{})

	if err != nil {
		log.Println("get chat for exit error:", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctrl.service.Exit(&chat)
}

func (ctrl *ChatCtrl) Clean(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if err != nil {
		log.Println("parse id for clean error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if status := ctrl.checkRights(id, r); status != 0 {
		w.WriteHeader(status)
		return
	}

	chat, err := ctrl.service.Get(id, []string{})

	if err != nil {
		log.Println("get chat for clean error:", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctrl.service.Exit(&chat)
}

func (ctrl *ChatCtrl) checkRights(id int64, r *http.Request) int {
	claims := getClaims(r)
	chat, err := ctrl.service.Get(id, []string{"users"})
	isExist := false

	if err != nil {
		return http.StatusNotFound
	}
	for _, user := range chat.Users {
		if user.ID == claims.Id {
			isExist = true
		}
	}
	if !isExist {
		return http.StatusForbidden
	}

	return 0
}

func (ctrl *ChatCtrl) getUnreadCount(chats []*ewc.Chat) []dao.ChatData {
	length := len(chats)
	chatData := make([]dao.ChatData, 0, length)
	chatIds := make([]int64, 0, length)
	chatMap := make(map[int64]*ewc.Chat, length)

	for i, chat := range chats {
		chatIds[i] = chat.ID
		chatMap[chat.ID] = chat
	}

	query := `
		select chats.id, count(messages.*)
		join on messages on messages.chat_id = chats.id
		where chats.id in (?) and messages.is_read = true
	`
	log.Println(query)

	return chatData
}
