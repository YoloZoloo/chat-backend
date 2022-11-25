package src

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "github.com/astaxie/session"
	_ "github.com/go-sql-driver/mysql"
)

type UserCredentials struct {
	UserName string `json:"user_id"`
	Password string `json:"password"`
}

type loginResponse struct {
	Name  string `json:"userName"`
	Token string `json:"token"`
	UID   uint   `json:"uniqID"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body)
	var creds UserCredentials
	err := json.Unmarshal(reqBody, &creds)
	if err != nil {
		respondError(w, "Can't parse request", http.StatusBadRequest)
		return
	}

	fmt.Println("user_id: " + creds.UserName)
	fmt.Println("password: " + creds.Password)

	user, err := CheckPassword(creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	} else {
		jwtToken, err := GenerateJWT(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respData, err := json.Marshal(&loginResponse{Token: jwtToken, UID: user.ID, Name: user.Firstname + " " + user.Lastname})
		if err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			AllowOriginAccess(w)
			w.Write([]byte(respData))
			return
		}
	}
}

