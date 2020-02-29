package controller

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"server/core/ewc"
	"server/model/dao"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

const friendCount = 10

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

	util := ewc.Util{}
	util.Setup(&ewc.SetupData{
		DbDriver:         driver,
		ConnectionString: connectionString,
	})

	db := getDb()
	db.AutoMigrate(&ewc.User{})
	db.AutoMigrate(&ewc.Friend{})

	for i := 0; i < chatCount; i++ {
		idx := strconv.Itoa(i)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password_"+idx), bcrypt.DefaultCost)
		resetHashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password_"+idx), bcrypt.DefaultCost)

		if i%2 == 0 {
			// isPersonal = true
		}

		user := ewc.User{
			Login:         "user_" + idx,
			Password:      string(hashedPassword),
			ResetPassword: string(resetHashedPassword),
		}
		db.Save(&user)

		// friends
		for j := 0; j < friendCount; j++ {
			db.Save(&ewc.Friend{
				UserID:   user.ID,
				FriendID: int64(j + 1),
			})
		}
	}

	db.Close()
	Config = cfg
}

func createMResponse(method string, addr string, vars map[string]string, rbody []byte, handler func(w http.ResponseWriter, r *http.Request)) (int, []byte) {
	token := createJwt()
	r := httptest.NewRequest(method, addr, bytes.NewReader(rbody))
	r = mux.SetURLVars(r, vars)
	r.Header.Add("X-Auth-Token", token)

	w := httptest.NewRecorder()

	handler(w, r)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	return resp.StatusCode, body
}

func TestLogin(t *testing.T) {
	setupUser()
	defer os.Remove(connectionString)

	ctrl := NewUserCtrl(cfg)
	data, _ := json.Marshal(map[string]string{
		"login":    "user_0",
		"password": "password_0",
	})
	status, body := createMResponse(http.MethodPost, "http://localhost/login", nil, data, ctrl.Login)
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
		"login":          "user_999",
		"password":       "password_999",
		"reset_password": "password_000",
	})
	status, body := createMResponse(http.MethodPost, "http://localhost/registration", nil, data, ctrl.Registration)
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
	ps := map[string]string{
		"id": "1",
	}
	status, body := createMResponse(http.MethodPost, "http://localhost", ps, nil, ctrl.RefreshToken)
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
	ps := map[string]string{
		"id": "1",
	}
	data, _ := json.Marshal(ewc.User{
		ID:    goodId,
		Login: "new login",
	})
	status, body := createMResponse(http.MethodPost, "http://localhost/users/1", ps, data, ctrl.Update)
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
	ps := map[string]string{
		"id": "1",
	}
	status, body := createMResponse(http.MethodPost, "http://localhost/users/1", ps, nil, ctrl.Get)
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
	ps := map[string]string{
		"login": "user_0",
	}
	status, body := createMResponse(http.MethodPost, "http://localhost/users", ps, nil, ctrl.GetByLogin)
	user := ewc.User{}

	assert.Equal(t, http.StatusOK, status)

	if err := json.Unmarshal(body, &user); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.Equal(t, "user_0", user.Login)
}

func TestGetFriends(t *testing.T) {
	setupUser()
	defer os.Remove(connectionString)

	ctrl := NewUserCtrl(cfg)
	status, body := createMResponse(http.MethodGet, "http://localhost/users/friends", nil, nil, ctrl.GetFriends)
	friends := make([]ewc.User, 0)

	assert.Equal(t, http.StatusOK, status)

	if err := json.Unmarshal(body, &friends); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.NotEmpty(t, friends)
}

func TestAddFriend(t *testing.T) {
	setupUser()
	defer os.Remove(connectionString)

	ctrl := NewUserCtrl(cfg)
	// create new user
	data, _ := json.Marshal(map[string]string{
		"login":          "user_999",
		"password":       "password_999",
		"reset_password": "password_000",
	})
	status, body := createMResponse(http.MethodPost, "http://localhost/registration", nil, data, ctrl.Registration)

	// add new user as friend
	data, _ = json.Marshal(map[string]string{
		"login": "user_999",
	})
	status, body = createMResponse(http.MethodPost, "http://localhost/users/friends", nil, data, ctrl.AddFriend)
	assert.Equal(t, http.StatusCreated, status)

	friend := ewc.User{}

	if err := json.Unmarshal(body, &friend); err != nil {
		assert.Failf(t, "invalid body: %s", string(body))
		return
	}

	assert.Equal(t, friend.Login, "user_999")
}

func TestDeleteFriend(t *testing.T) {
	setupUser()
	defer os.Remove(connectionString)

	ctrl := NewUserCtrl(cfg)
	ps := map[string]string{
		"id": "9",
	}
	status, _ := createMResponse(http.MethodPost, "http://localhost/users/friends/9", ps, nil, ctrl.DeleteFriend)
	assert.Equal(t, http.StatusOK, status)
}
