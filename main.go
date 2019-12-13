package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"server/controller"
	"server/middleware"
	"server/model/dao"

	"github.com/julienschmidt/httprouter"
)

const defaultConfigPath = "./cfg.json"

var config *dao.Config

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

type httpHandler = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params)

func jwtHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, handler httpHandler) {
	if err := middleware.TokenValidation(w, r, ps); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	handler(w, r, ps)
}

func main() {
	userCtrl := controller.NewUserCtrl(config)
	chatCtrl := controller.NewChatCtrl(config)
	messageCtrl := controller.NewMessageCtrl(config)

	router := httprouter.New()

	// user
	router.POST("/login", userCtrl.Login)
	router.POST("/registration", userCtrl.Registration)
	router.PUT("/users/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, userCtrl.Update)
	})
	router.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, userCtrl.Get)
	})

	// chat
	router.GET("/chats", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, chatCtrl.GetList)
	})
	router.GET("/chats/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, chatCtrl.Get)
	})
	router.POST("/chats", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, chatCtrl.Create)
	})
	router.DELETE("/chats/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, chatCtrl.Delete)
	})
	router.DELETE("/chats/:id/exit", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, chatCtrl.Exit)
	})
	router.DELETE("/chats/:id/clean", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, chatCtrl.Clean)
	})

	// message
	router.POST("/messages", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, messageCtrl.Create)
	})
	router.DELETE("/messages/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, messageCtrl.Delete)
	})

	log.Println("Server start on", config.ServiceAddress)

	if err := http.ListenAndServe(config.ServiceAddress, router); err != nil {
		panic("start server error: " + err.Error())
	}
}
