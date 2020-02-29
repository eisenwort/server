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
	"github.com/julienschmidt/httprouter"
)

const defaultConfigPath = "./cfg.json"

var config *dao.Config

type httpHandler = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
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

func jwtHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, handler httpHandler) {
	if err := middleware.TokenValidation(w, r, ps); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	handler(w, r, ps)
}

func mjwtHandler(w http.ResponseWriter, r *http.Request, handler mhttpHandler) {
	if err := middleware.MuxTokenValidation(w, r); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	handler(w, r)
}

func muxRouter() *mux.Router {
	userCtrl := controller.NewUserCtrl(config)
	chatCtrl := controller.NewChatCtrl(config)
	messageCtrl := controller.NewMessageCtrl(config)
	router := mux.NewRouter()

	// user
	router.HandleFunc("/login", userCtrl.Login).Methods(http.MethodPost)
	router.HandleFunc("registration", userCtrl.Registration).Methods(http.MethodPost)
	router.HandleFunc("users/:id/refresh", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, userCtrl.RefreshToken)
	}).Methods(http.MethodPost)
	router.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, userCtrl.Update)
	}).Methods(http.MethodPut)
	router.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, userCtrl.Get)
	}).Methods(http.MethodGet)
	router.HandleFunc("/users/friends", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, userCtrl.GetFriends)
	}).Methods(http.MethodGet)
	router.HandleFunc("/users/friends", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, userCtrl.AddFriend)
	}).Methods(http.MethodPost)
	router.HandleFunc("/users/friends/{id}", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, userCtrl.DeleteFriend)
	}).Methods(http.MethodDelete)
	router.HandleFunc("/users/login/{login}", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, userCtrl.GetByLogin)
	}).Methods(http.MethodGet)

	// chat
	router.HandleFunc("/chats", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, chatCtrl.GetList)
	}).Methods(http.MethodGet)
	router.HandleFunc("/chats/{id}", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, chatCtrl.Get)
	}).Methods(http.MethodGet)
	router.HandleFunc("/chats", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, chatCtrl.Create)
	}).Methods(http.MethodPost)
	router.HandleFunc("/chats/{id}", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, chatCtrl.Delete)
	}).Methods(http.MethodDelete)
	router.HandleFunc("/chats/{id}/exit", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, chatCtrl.Exit)
	}).Methods(http.MethodDelete)
	router.HandleFunc("/chats/{id}/clean", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, chatCtrl.Clean)
	}).Methods(http.MethodDelete)

	// message
	router.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, messageCtrl.Create)
	}).Methods(http.MethodPost)
	router.HandleFunc("/messages/{id}", func(w http.ResponseWriter, r *http.Request) {
		mjwtHandler(w, r, messageCtrl.Delete)
	}).Methods(http.MethodDelete)

	return router
}

/*func createRouter() *httprouter.Router {
	userCtrl := controller.NewUserCtrl(config)
	chatCtrl := controller.NewChatCtrl(config)
	messageCtrl := controller.NewMessageCtrl(config)
	router := httprouter.New()

	// user
	router.POST("/login", userCtrl.Login)
	router.POST("/registration", userCtrl.Registration)
	router.POST("/users/:id/refresh", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, userCtrl.RefreshToken)
	})
	router.PUT("/users/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, userCtrl.Update)
	})
	router.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, userCtrl.Get)
	})
	router.GET("/users/:login/friend", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		jwtHandler(w, r, ps, userCtrl.GetByLogin)
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

	return router
}
*/
func main() {
	util := ewc.NewUtil()
	util.Setup(&ewc.SetupData{
		DbDriver:         config.Driver,
		ConnectionString: config.ConnectionString,
	})
	defer util.CloseApp()
	middleware.Setup(config)

	router := muxRouter()
	log.Println("Server start on", config.ServiceAddress)

	if err := http.ListenAndServe(config.ServiceAddress, router); err != nil {
		panic("start server error: " + err.Error())
	}
}
