package src

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func StartServer() {
	hub := newHub()
	go hub.run()
	r := mux.NewRouter()
	r.Methods(http.MethodOptions).HandlerFunc(RespondOptions)
	r.Path("/api/login").Methods(http.MethodPost).HandlerFunc(Login)
	r.Path("/api/register").Methods(http.MethodPost).HandlerFunc(RegisterUser)
	r.Path("/ws").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ListenWS(hub, w, r)
	})
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

func AllowOriginAccess(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}

func respondError(w http.ResponseWriter, err string, status int) {
	AllowOriginAccess(w)
	http.Error(w, err, status)
}

func ShowDefaultIcon(w http.ResponseWriter, r *http.Request) {
	buf, err := os.ReadFile("common/defaultIcon.png")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "image/png")
	AllowOriginAccess(w)
	w.Write(buf)
}

func RespondOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "OPTIONS, GET, HEAD, POST")
	AllowOriginAccess(w)
	w.WriteHeader(http.StatusOK)
}

func SetEnvVariables() {
	envFile, err := os.Open("./.env")
	if err != nil {
		log.Fatal(err)
	}
	defer envFile.Close()

	var envVariables []string
	scanner := bufio.NewScanner(envFile)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		envVariables = append(envVariables, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	db_username := envVariables[0]
	db_password := envVariables[1]
	db_host := envVariables[2]
	db_port := envVariables[3]
	db_database := envVariables[4]
	jwt_sign_key := envVariables[5]

	os.Setenv("GO_CHAT_DB_USERNAME", db_username)
	os.Setenv("GO_CHAT_DB_PASSWORD", db_password)
	os.Setenv("GO_CHAT_DB_HOST", db_host)
	os.Setenv("GO_CHAT_DB_PORT", db_port)
	os.Setenv("GO_CHAT_DATABASE", db_database)
	os.Setenv("GO_CHAT_JWT_KEY", jwt_sign_key)
}
