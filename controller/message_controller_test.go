package controller

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"server/core/ewc"
	"strconv"
	"testing"
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

const msgCount = 30

func setupMsg() {
	err := godotenv.Load()

	if err != nil {
		log.Println(err)
	}

	driver = os.Getenv("DB_DRIVER")
	connectionString = "msg_test.sqlite"
	cfg.Driver = driver
	cfg.ConnectionString = connectionString

	ewc.Setup(ewc.SetupData{
		DbDriver:         driver,
		ConnectionString: connectionString,
	})

	db := getDb()
	db.AutoMigrate(&ewc.Message{})
	db.AutoMigrate(&ewc.Chat{})
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
		db.Save(&chat)

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

func TestCreateMessage(t *testing.T) {
	setupChats()
	defer os.Remove(connectionString)

	ctrl := NewMessageCtrl(cfg)
	message := ewc.Message{
		UserID: goodId,
		ChatID: goodId,
		Text:   "msg text",
	}
	body, _ := json.Marshal(message)
	status, _ := createResponse(http.MethodPost, "http://localhost/messages", nil, bytes.NewReader(body), ctrl.Create)

	assert.Equal(t, http.StatusCreated, status)
}

func TestDeleteMessage(t *testing.T) {
	setupChats()
	defer os.Remove(connectionString)

	ctrl := NewMessageCtrl(cfg)
	ps := []httprouter.Param{
		httprouter.Param{
			Key:   "id",
			Value: "1",
		},
	}
	message := ewc.Message{
		ID:     goodId,
		UserID: goodId,
		ChatID: goodId,
		Text:   "msg text",
	}
	body, _ := json.Marshal(message)
	status, _ := createResponse(http.MethodPost, "http://localhost/messages/1", ps, bytes.NewReader(body), ctrl.Delete)

	assert.Equal(t, http.StatusOK, status)
}
