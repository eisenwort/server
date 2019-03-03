package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"server/controller"
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
}

func main() {
	userCtrl := controller.NewUserCtrl(config)
	chatCtrl := controller.NewChatCtrl(config)
	messageCtrl := controller.NewMessageCtrl(config)

	router := httprouter.New()
	router.POST("/login", userCtrl.Login)
	router.POST("/registration", userCtrl.Registration)
	router.PUT("/users/:id", userCtrl.Update)

	router.GET("/chats", chatCtrl.GetList)
	router.GET("/chats/:id", chatCtrl.Get)
	router.POST("/chats", chatCtrl.Create)
	router.DELETE("/chats/:id", chatCtrl.Delete)
	router.DELETE("/chats/:id/exit", chatCtrl.Exit)
	router.DELETE("/chats/:id/clean", chatCtrl.Clean)

	router.POST("/messages", messageCtrl.Create)
	router.DELETE("/messages/:id", messageCtrl.Delete)

	log.Println("Server start on", config.ServiceAddress)

	if err := http.ListenAndServe(config.ServiceAddress, router); err != nil {
		panic("start server error: " + err.Error())
	}
}
