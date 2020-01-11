package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"server/core/ewc"
	"server/model/dao"
	"strconv"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

var cfg = &dao.Config{
	Driver:           "sqlite3",
	ConnectionString: "test.sqlite",
	JwtSign:          "123456",
	PageLimit:        pageCount,
}

const chatCount = 30
const pageCount = 20

var goodId = int64(1)
var driver = ""
var connectionString = ""

func setupChats() {
	err := godotenv.Load()

	if err != nil {
		log.Println(err)
	}

	driver = os.Getenv("DB_DRIVER")
	connectionString = "chat_test.sqlite"
	cfg.Driver = driver
	cfg.ConnectionString = connectionString

	ewc.Setup(ewc.SetupData{
		DbDriver:         driver,
		ConnectionString: connectionString,
	})

	db := getDb()
	db.AutoMigrate(&ewc.Message{})
	db.AutoMigrate(&ewc.User{})
	db.AutoMigrate(&ewc.Chat{})
	db.AutoMigrate(&ewc.ChatUser{})
	expiredTime := time.Now().Add(1 * time.Hour)
	now := time.Now()

	for i := 0; i < chatCount; i++ {
		idx := strconv.Itoa(i)
		isPersonal := false

		if i%2 == 0 {
			isPersonal = true
		}

		chat := ewc.Chat{
			OwnerID:        goodId,
			UnreadMessages: 0,
			Name:           "chat_" + idx,
			Personal:       isPersonal,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		user := ewc.User{
			Login: "login_" + idx,
		}
		db.Save(&chat)
		db.Save(&user)

		chatUser := ewc.ChatUser{
			ChatID: chat.ID,
			UserID: user.ID,
		}
		db.Save(&chatUser)

		for j := 0; j < chatCount; j++ {
			msg := ewc.Message{
				UserID:    goodId,
				ChatID:    chat.ID,
				Text:      "msg_text_" + idx,
				CreatedAt: now,
				UpdatedAt: now,
				ExpiredAt: expiredTime,
			}
			db.Save(&msg)
		}
	}

	db.Close()
	Config = cfg
}

func getDb() *gorm.DB {
	db, err := gorm.Open(driver, connectionString)

	if err != nil {
		log.Println("open db error:", err.Error())
		return nil
	}

	return db
}

func createJwt() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, ewc.JwtClaims{
		Id: goodId,
	})
	tokenString, err := token.SignedString([]byte(cfg.JwtSign))

	if err != nil {
		log.Println("create token error", err.Error())
		return ""
	}

	return tokenString
}

func createResponse(method string, addr string, ps httprouter.Params, rbody io.Reader, handler func(w http.ResponseWriter, r *http.Request, ps httprouter.Params)) (int, []byte) {
	token := createJwt()
	r := httptest.NewRequest(method, addr, rbody)
	r.Header.Add("X-Auth-Token", token)
	w := httptest.NewRecorder()

	handler(w, r, ps)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	return resp.StatusCode, body
}

func TestGetList(t *testing.T) {
	setupChats()
	defer os.Remove(connectionString)

	ctrl := NewChatCtrl(cfg)
	status, body := createResponse(http.MethodGet, "http://localhost/chats", nil, nil, ctrl.GetList)
	chats := make([]*ewc.Chat, 0)

	assert.Equal(t, http.StatusOK, status)

	if err := json.Unmarshal(body, &chats); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.Equal(t, chatCount, len(chats))

	for _, chat := range chats {
		assert.Equal(t, goodId, chat.OwnerID)
		// assert.NotNil(t, chat.Users)
		// assert.NotEmpty(t, chat.Users)
	}
}

func TestGet(t *testing.T) {
	setupChats()
	defer os.Remove(connectionString)

	ctrl := NewChatCtrl(cfg)
	ps := []httprouter.Param{
		httprouter.Param{
			Key:   "id",
			Value: "1",
		},
	}
	status, body := createResponse(http.MethodGet, "http://localhost/chats/1", ps, nil, ctrl.Get)
	chat := ewc.Chat{}

	assert.Equal(t, http.StatusOK, status)

	if err := json.Unmarshal(body, &chat); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.Equal(t, goodId, chat.ID)

	// not found
	ps = []httprouter.Param{
		httprouter.Param{
			Key:   "id",
			Value: "111",
		},
	}
	status, body = createResponse(http.MethodGet, "http://localhost/chats/1111", ps, nil, ctrl.Get)
	assert.Equal(t, http.StatusNotFound, status)

	// forbidden
}

func TestCreate(t *testing.T) {
	setupChats()
	defer os.Remove(connectionString)

	ctrl := NewChatCtrl(cfg)
	body, _ := json.Marshal(ewc.Chat{
		OwnerID: goodId,
		Name:    "new chat",
		Users: []ewc.User{
			ewc.User{
				ID: goodId,
			},
		},
	})
	status, body := createResponse(http.MethodPost, "http://localhost/chats", nil, bytes.NewReader(body), ctrl.Create)
	chat := ewc.Chat{}

	assert.Equal(t, http.StatusCreated, status)

	if err := json.Unmarshal(body, &chat); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.NotEqual(t, 0, chat.ID)
	assert.Equal(t, false, chat.Personal)
	assert.Equal(t, "new chat", chat.Name)
	assert.Equal(t, goodId, chat.OwnerID)

	ps := []httprouter.Param{
		httprouter.Param{
			Key:   "id",
			Value: fmt.Sprintf("%d", chat.ID),
		},
		httprouter.Param{
			Key:   "include",
			Value: "users",
		},
	}
	status, body = createResponse(http.MethodGet, "http://localhost/chats/1111", ps, nil, ctrl.Get)
	assert.Equal(t, http.StatusOK, status)

	for _, user := range chat.Users {
		assert.Equal(t, goodId, user.ID)
	}
}

func TestDelete(t *testing.T) {
	setupChats()
	defer os.Remove(connectionString)

	ctrl := NewChatCtrl(cfg)
	ps := []httprouter.Param{
		httprouter.Param{
			Key:   "id",
			Value: "1",
		},
	}
	status, _ := createResponse(http.MethodDelete, "http://localhost/chats/1", ps, nil, ctrl.Delete)
	assert.Equal(t, http.StatusOK, status)

	status, _ = createResponse(http.MethodGet, "http://localhost/chats/1", ps, nil, ctrl.Get)
	assert.Equal(t, http.StatusNotFound, status)
}
