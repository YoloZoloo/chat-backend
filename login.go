package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
	queryString := `select IFNULL(firstname, "") as firstname, IFNULL(lastname, "") as lastname from user_m where id = ?`

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
	json.Unmarshal(reqBody, &data)

	user_id := data.UserID
	password := data.Password

	fmt.Println("user_id: " + user_id)
	fmt.Println("password: " + password)

	uniqueID, err := CheckPassword(user_id, password)

	if err != nil {
		http.Error(w, "Couldn't generate token.", http.StatusServiceUnavailable)
		return
	} else {
		jwtToken, err := GenerateJWT(user_id, uniqueID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		name, err := GetNameOfTheUser(uniqueID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		respData, err := json.Marshal(&loginResponse{Token: jwtToken, UID: uniqueID, Name: name})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		} else {
			AllowOriginAccess(w, r)
			w.Write([]byte(respData))
			return
		}
	}
}

//This function should be replaced by your own DB. It doesn't how to be MYSQL.
func OpenDB() (*sql.DB, error) {
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
	if err != nil {
		panic(err.Error())
	}
	db, err := sql.Open("mysql", envVariables[0]+"@tcp("+envVariables[1]+")/"+envVariables[2])
	return db, err
}
