package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"server/controller"
	"server/core/ewc"
	"server/middleware"
	"server/model/dao"

	"github.com/gorilla/mux"
)

const defaultConfigPath = "./cfg.json"

var config *dao.Config

type mhttpHandler = func(w http.ResponseWriter, r *http.Request)

func init() {
	pathPtr := flag.String("config", defaultConfigPath, "Path for configuration file")
	flag.Parse()

	configBytes, err := ioutil.ReadFile(*pathPtr)

	if err != nil {
		panic("Read config file error")
	}

	config = new(dao.Config)

	if err = json.Unmarshal(configBytes, config); err != nil {
		panic("unmarshal config file error: " + err.Error())
	}

	controller.Config = config
}

func jwtHandler(w http.ResponseWriter, r *http.Request, handler mhttpHandler) {
	if err := middleware.MuxTokenValidation(w, r); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	handler(w, r)
}

func createRouter() http.Handler {
	userCtrl := controller.NewUserCtrl(config)
	chatCtrl := controller.NewChatCtrl(config)
	messageCtrl := controller.NewMessageCtrl(config)
	router := mux.NewRouter()

	// user
	router.HandleFunc("/login", userCtrl.Login).Methods(http.MethodPost)
	router.HandleFunc("registration", userCtrl.Registration).Methods(http.MethodPost)
	router.HandleFunc("users/:id/refresh", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, userCtrl.RefreshToken)
	}).Methods(http.MethodPost)
	router.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, userCtrl.Update)
	}).Methods(http.MethodPut)
	router.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, userCtrl.Get)
	}).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}/friends", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, userCtrl.GetFriends)
	}).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}/friends", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, userCtrl.AddFriend)
	}).Methods(http.MethodPost)
	router.HandleFunc("/users/{user_id}/friends/{id}", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, userCtrl.DeleteFriend)
	}).Methods(http.MethodDelete)
	router.HandleFunc("/users/login/{login}", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, userCtrl.GetByLogin)
	}).Methods(http.MethodGet)

	// chat
	router.HandleFunc("/chats", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, chatCtrl.GetList)
	}).Methods(http.MethodGet)
	router.HandleFunc("/chats/{id}", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, chatCtrl.Get)
	}).Methods(http.MethodGet)
	router.HandleFunc("/chats", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, chatCtrl.Create)
	}).Methods(http.MethodPost)
	router.HandleFunc("/chats/{id}", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, chatCtrl.Delete)
	}).Methods(http.MethodDelete)
	router.HandleFunc("/chats/{id}/exit", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, chatCtrl.Exit)
	}).Methods(http.MethodDelete)
	router.HandleFunc("/chats/{id}/clean", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, chatCtrl.Clean)
	}).Methods(http.MethodDelete)
	router.HandleFunc("/chats/{chat_id}", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, messageCtrl.GetLastId)
	}).Methods(http.MethodHead)

	// message
	router.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, messageCtrl.Create)
	}).Methods(http.MethodPost)
	router.HandleFunc("/messages/{id}", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, messageCtrl.Delete)
	}).Methods(http.MethodDelete)
	router.HandleFunc("/chats/{chat_id}/messages", func(w http.ResponseWriter, r *http.Request) {
		jwtHandler(w, r, messageCtrl.GetByChat)
	}).Methods(http.MethodGet)

	return router
}

func main() {
	util := ewc.NewUtil()
	util.Setup(&ewc.SetupData{
		DbDriver:         config.Driver,
		ConnectionString: config.ConnectionString,
	})
	defer util.CloseApp()
	middleware.Setup(config)

	router := createRouter()
	log.Println("Server start on", config.ServiceAddress)

	if err := http.ListenAndServe(config.ServiceAddress, router); err != nil {
		panic("start server error: " + err.Error())
	}
}
