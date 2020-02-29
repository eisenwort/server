package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"server/core/ewc"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
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

	util := ewc.Util{}
	util.Setup(&ewc.SetupData{
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
	status, _ := createMResponse(http.MethodPost, "http://localhost/messages", nil, body, ctrl.Create)

	assert.Equal(t, http.StatusCreated, status)
}

func TestDeleteMessage(t *testing.T) {
	setupChats()
	defer os.Remove(connectionString)

	ctrl := NewMessageCtrl(cfg)
	ps := map[string]string{
		"id": "1",
	}
	message := ewc.Message{
		ID:     goodId,
		UserID: goodId,
		ChatID: goodId,
		Text:   "msg text",
	}
	body, _ := json.Marshal(message)
	status, _ := createMResponse(http.MethodPost, "http://localhost/messages/1", ps, body, ctrl.Delete)

	assert.Equal(t, http.StatusOK, status)
}
