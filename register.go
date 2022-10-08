package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func GetUniqID(userID string) (int, error) {
	db, err := OpenDB()
	if err != nil {
		return 0, err
	}

	var uniqID int
	queryString := "SELECT id FROM chat.user_m where user_id = ?"
	rows, err := db.Query(queryString, userID)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&uniqID); err != nil {
			fmt.Println(err)
			return 0, err
		}
	}
	rows.Close()

	return uniqID, nil
}

func GetAdminPrivilige(userID string) int {
	db, err := OpenDB()
	if err != nil {
		return 0
	}

	queryString := " SELECT count(*) FROM chat.user_m where user_id = ?"
	rows, err := db.Query(queryString, userID)
	if err != nil {
		return 0
	}
	defer rows.Close()

	var count int
	if err := rows.Scan(&count); err != nil {
		if err != nil {
			return 0
		}
	}
	rows.Close()
	return count
}

type registerForm struct {
	UserID    string `json:"user_id"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondError(w, "Reading requets body failed", http.StatusBadRequest)
		return
	}

	var data registerForm
	err = json.Unmarshal(reqBody, &data)
	if err != nil {
		respondError(w, "Can't parse request", http.StatusBadRequest)
		return
	}

	user_id := data.UserID
	password := data.Password
	firstName := data.FirstName
	lastName := data.LastName

	adminAuthority := GetAdminPrivilige(user_id)
	fmt.Println(user_id, password, firstName, lastName, adminAuthority)

	db, err := OpenDB()
	queryString := "insert into chat.user_m(user_id, password, firstname, lastname, adminAuthority) values (?, ?, ?, ?, ?)"
	stmt, err := db.Prepare(queryString)
	if err != nil {
		respondError(w, "couldn't prepare insertadfa statement.", http.StatusBadRequest)
		return
	}
	defer stmt.Close()

	hashedPassword, _ := HashPassword(password)
	_, err = stmt.Exec(user_id, hashedPassword, firstName, lastName, adminAuthority)
	if err != nil {
		respondError(w, "insert statement failed", http.StatusBadRequest)
		return
	}
	stmt.Close()

	//Get uniqID
	uniqID, err := GetUniqID(user_id)
	fmt.Println("uniqID: ", uniqID)
	if err != nil {
		respondError(w, "couldn't get uniqID.", http.StatusBadRequest)
		return
	}

	//Insert to lobby
	queryString = "INSERT INTO chat.grouproom_m (grouproom_id,guest_id) VALUES (0,?)"
	stmt, err = db.Prepare(queryString)
	if err != nil {
		respondError(w, "couldn't prepare insert statement.", http.StatusBadRequest)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(uniqID)
	if err != nil {
		respondError(w, "insert statement for lobby failed", http.StatusBadRequest)
		return
	}
	stmt.Close()

	queryString = "SELECT id FROM chat.user_m where id != ? ORDER BY id ASC"
	rows, err := db.Query(queryString, uniqID)

	if err != nil {
		respondError(w, "select statement failed", http.StatusBadRequest)
		return
	}
	defer rows.Close()
	privChatInsertsValues := ""
	uniqIDStr := strconv.Itoa(uniqID)

	for rows.Next() {
		var insertValue string
		if err := rows.Scan(&insertValue); err != nil {
			respondError(w, "Scanning rows failed", http.StatusBadRequest)
			return
		}

		fmt.Println("insertValue: ", insertValue)
		privChatInsertsValues = privChatInsertsValues + "(" + uniqIDStr + "," + insertValue + "),"
	}
	rows.Close()

	fmt.Println("privChatInsertsValues: ", privChatInsertsValues)
	sz := len(privChatInsertsValues)
	if sz > 0 {
		privChatInsertsValues = privChatInsertsValues[:sz-1] + ";"
		queryString = "INSERT INTO chat.privateroom_m (idA,idB) VALUES " + privChatInsertsValues
		stmt, err = db.Prepare(queryString)
		if err != nil {
			respondError(w, "couldn't prepare insert statement.", http.StatusBadRequest)
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec()
		if err != nil {
			respondError(w, "insert statement failed", http.StatusBadRequest)
			return
		}
		stmt.Close()
	}

	AllowOriginAccess(w)
	w.WriteHeader(http.StatusOK) // 200 OK
	return
}
