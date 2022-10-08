package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	_ "github.com/astaxie/session"
	_ "github.com/go-sql-driver/mysql"
)

type userCredentials struct {
	ID       int    `json:"id"`
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

type loginResponse struct {
	Name  string `json:"userName"`
	Token string `json:"token"`
	UID   int16  `json:"uniqID"`
}

func GetNameOfTheUser(uniqID int16) (string, error) {
	db, err := OpenDB()
	if err != nil {
		return "", err
	}
	queryString := `select IFNULL(firstname, "") as firstname, IFNULL(lastname, "") as lastname from chat.user_m where id = ?`

	rows, err := db.Query(queryString, uniqID)

	if err != nil {
		return "", err
	}
	defer rows.Close()

	var firstName string
	var lastName string
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		if err := rows.Scan(&firstName, &lastName); err != nil {
			if err != nil {
				return "", err
			}
		}
	}
	rows.Close()
	return firstName + " " + lastName, nil
}

func Login(w http.ResponseWriter, r *http.Request) {

	reqBody, _ := ioutil.ReadAll(r.Body)
	var data userCredentials
	err := json.Unmarshal(reqBody, &data)
	if err != nil {
		respondError(w, "Can't parse request", http.StatusBadRequest)
		return
	}
	user_id := data.UserID
	password := data.Password

	fmt.Println("user_id: " + user_id)
	fmt.Println("password: " + password)

	uniqueID, err := CheckPassword(user_id, password)

	if err != nil {
		http.Error(w, "Wrong credentials", http.StatusUnauthorized)
		return
	} else {
		jwtToken, err := GenerateJWT(user_id, uniqueID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		name, err := GetNameOfTheUser(uniqueID)
		if err != nil {
			respondError(w, err.Error(), http.StatusExpectationFailed)
			return
		}

		respData, err := json.Marshal(&loginResponse{Token: jwtToken, UID: uniqueID, Name: name})
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
