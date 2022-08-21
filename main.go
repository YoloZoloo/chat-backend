package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	_ "github.com/astaxie/session"
	_ "github.com/astaxie/session/providers/memory"
	_ "github.com/go-sql-driver/mysql"
)

func ShowDefaultIcon(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadFile("common/defaultIcon.png")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "image/png")
	AllowOriginAccess(w, r)
	w.Write(buf)
}

func RespondOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "OPTIONS, GET, HEAD, POST")
	AllowOriginAccess(w, r)
	w.WriteHeader(http.StatusOK)
}
func AllowOriginAccess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}

func main() {
	r := mux.NewRouter()
	r.Methods(http.MethodOptions).HandlerFunc(RespondOptions)
	r.Path("/api/login").Methods(http.MethodPost).HandlerFunc(Login)
	r.Path("/api/register").Methods(http.MethodPost).HandlerFunc(RegisterUser)
	r.HandleFunc("/defaultIcon.png", ShowDefaultIcon)

	s := r.PathPrefix("/api/chatscreen").Subrouter()
	s.Path("/getChatRooms").Methods(http.MethodGet).HandlerFunc(GetSubscribedRooms)
	s.Path("/getGroupChat").Methods(http.MethodPost).HandlerFunc(GetGroupChat)
	s.Path("/getPrivateChat").Methods(http.MethodPost).HandlerFunc(GetPrivateChat)

	fmt.Printf("Starting server at port 9999\n")

	if err := http.ListenAndServe(":9999", r); err != nil {
		log.Fatal(err)
	}
}
