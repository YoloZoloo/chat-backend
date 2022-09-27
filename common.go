package main

import (
	"database/sql"
	"net/http"
	"os"
)

func AllowOriginAccess(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}

func respondError(w http.ResponseWriter, err string, status int) {
	AllowOriginAccess(w)
	http.Error(w, err, status)
}

//This function should be replaced by your own DB. It doesn't how to be MYSQL.
func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("mysql",
		os.Getenv("GO_CHAT_DB_USERNAME")+":"+os.Getenv("GO_CHAT_DB_PASSWORD")+
			"@tcp("+
			os.Getenv("GO_CHAT_DB_HOST")+":"+os.Getenv("GO_CHAT_DB_PORT")+
			")/"+
			os.Getenv("GO_CHAT_DATABASE"))
	return db, err
}
