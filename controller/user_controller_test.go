package controller

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"server/core/ewc"
	"server/model/dao"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func setupUser() {
	err := godotenv.Load()

	if err != nil {
		log.Println(err)
	}

	driver = os.Getenv("DB_DRIVER")
	// connectionString = os.Getenv("DB")
	connectionString = "test_user.sqlite"
	cfg.Driver = driver
	cfg.ConnectionString = connectionString

	ewc.Setup(ewc.SetupData{
		DbDriver:         driver,
		ConnectionString: connectionString,
	})

	db := getDb()
	db.AutoMigrate(&ewc.User{})

	for i := 0; i < chatCount; i++ {
		idx := strconv.Itoa(i)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password_"+idx), bcrypt.DefaultCost)
		resetHashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password_"+idx), bcrypt.DefaultCost)

		if i%2 == 0 {
			// isPersonal = true
		}

		db.Save(&ewc.User{
			Login:         "user_" + idx,
			Password:      string(hashedPassword),
			ResetPassword: string(resetHashedPassword),
		})
	}

	db.Close()
	Config = cfg
}

func TestLogin(t *testing.T) {
	setupUser()
	defer os.Remove(connectionString)

	ctrl := NewUserCtrl(cfg)
	data, _ := json.Marshal(map[string]string{
		"login":    "user_0",
		"password": "password_0",
	})
	status, body := createResponse(http.MethodPost, "http://localhost/login", nil, bytes.NewReader(data), ctrl.Login)
	tokenData := dao.AuthData{}

	assert.Equal(t, http.StatusOK, status)

	if err := json.Unmarshal(body, &tokenData); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.NotEmpty(t, tokenData.Token)
	assert.NotEmpty(t, tokenData.RefreshToken)

}

func TestRegistration(t *testing.T) {
	setupUser()
	defer os.Remove(connectionString)

	ctrl := NewUserCtrl(cfg)
	data, _ := json.Marshal(map[string]string{
		"login":    "user_999",
		"password": "password_999",
	})
	status, body := createResponse(http.MethodPost, "http://localhost/registration", nil, bytes.NewReader(data), ctrl.Registration)
	tokenData := dao.AuthData{}

	assert.Equal(t, http.StatusCreated, status)

	if err := json.Unmarshal(body, &tokenData); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.NotEmpty(t, tokenData.Token)
	assert.NotEmpty(t, tokenData.RefreshToken)
}

func TestRefreshToken(t *testing.T) {
	setupUser()
	defer os.Remove(connectionString)

	ctrl := NewUserCtrl(cfg)
	ps := []httprouter.Param{
		httprouter.Param{
			Key:   "id",
			Value: "1",
		},
	}
	status, body := createResponse(http.MethodPost, "http://localhost", ps, nil, ctrl.RefreshToken)
	tokenData := dao.AuthData{}

	assert.Equal(t, http.StatusOK, status)

	if err := json.Unmarshal(body, &tokenData); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.NotEmpty(t, tokenData.Token)
	assert.NotEmpty(t, tokenData.RefreshToken)
}

func TestUpdate(t *testing.T) {
	setupUser()
	defer os.Remove(connectionString)

	ctrl := NewUserCtrl(cfg)
	ps := []httprouter.Param{
		httprouter.Param{
			Key:   "id",
			Value: "1",
		},
	}
	data, _ := json.Marshal(ewc.User{
		ID:    goodId,
		Login: "new login",
	})
	status, body := createResponse(http.MethodPost, "http://localhost/users/1", ps, bytes.NewReader(data), ctrl.Update)
	user := ewc.User{}

	assert.Equal(t, http.StatusOK, status)

	if err := json.Unmarshal(body, &user); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.Equal(t, goodId, user.ID)
	assert.Equal(t, "new login", user.Login)
}

func TestGetUser(t *testing.T) {
	setupUser()
	defer os.Remove(connectionString)

	ctrl := NewUserCtrl(cfg)
	ps := []httprouter.Param{
		httprouter.Param{
			Key:   "id",
			Value: "1",
		},
	}
	status, body := createResponse(http.MethodPost, "http://localhost/users/1", ps, nil, ctrl.Get)
	user := ewc.User{}

	assert.Equal(t, http.StatusOK, status)

	if err := json.Unmarshal(body, &user); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.Equal(t, goodId, user.ID)
}

func TestGetByLogin(t *testing.T) {
	setupUser()
	defer os.Remove(connectionString)

	ctrl := NewUserCtrl(cfg)
	ps := []httprouter.Param{
		httprouter.Param{
			Key:   "login",
			Value: "user_0",
		},
	}
	status, body := createResponse(http.MethodPost, "http://localhost/users", ps, nil, ctrl.GetByLogin)
	user := ewc.User{}

	assert.Equal(t, http.StatusOK, status)

	if err := json.Unmarshal(body, &user); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.Equal(t, "user_0", user.Login)
}
