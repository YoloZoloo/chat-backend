package src

import (
	model "chat-backend/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetUniqID(userID string) (int, error) {

	db, err := model.DbInit()

	if err != nil {
		return 0, err
	}
	var creds model.User
	result := db.First(&creds, "username = ?", userID)
	if result.RowsAffected == 0 || result.Error != nil {
		return 0, result.Error
	}
	// var uniqID int
	// queryString := "SELECT id FROM chat.user_m where user_id = ?"
	// rows, err := db.Query(queryString, userID)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return 0, err
	// }

	// defer rows.Close()
	// for rows.Next() {
	// 	if err := rows.Scan(&uniqID); err != nil {
	// 		fmt.Println(err)
	// 		return 0, err
	// 	}
	// }
	// rows.Close()

	// return uniqID, nil
	return 0, nil
}

func GetAdminPrivilige(userID string) int {
	db, err := model.DbInit()

	if err != nil {
		return 0
	}
	var creds model.User
	result := db.First(&creds, "username = ?", userID)
	if result.RowsAffected == 0 || result.Error != nil {
		return 0
	}
	return 0
	// queryString := " SELECT count(*) FROM chat.user_m where user_id = ?"
	// rows, err := db.Query(queryString, userID)
	// if err != nil {
	// 	return 0
	// }
	// defer rows.Close()

	// var count int
	// if err := rows.Scan(&count); err != nil {
	// 	if err != nil {
	// 		return 0
	// 	}
	// }
	// rows.Close()
	// return count
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

	adminAuthority := GetAdminPrivilige(data.UserID)
	fmt.Println(data.UserID, data.Password, data.FirstName, data.LastName, adminAuthority)

	db, err := model.DbInit()
	if err != nil {
		respondError(w, "", http.StatusInternalServerError)
		return
	}

	hashedPassword, _ := HashPassword(data.Password)
	user := model.User{
		Username:  data.UserID,
		Password:  hashedPassword,
		Firstname: data.FirstName,
		Lastname:  data.LastName,
	}
	result := db.Create(&user)
	if result.RowsAffected == 0 || result.Error != nil {
		respondError(w, "couldn't prepare insertadfa statement.", http.StatusBadRequest)
		return
	}

	var me model.User
	// result = map[string]interface{}{}
	// db.Model(&model.User{}).Take(&result, "username = ?", data.UserID)
	result = db.Take(&me, "username = ?", data.UserID)
	if result.RowsAffected == 0 || result.Error != nil {
		respondError(w, "couldn't create a user", http.StatusBadRequest)
		return
	}
	// result.QueryFields
	fmt.Println(result.Scan(""))
	fmt.Println(me)

	// insert to lobby
	room := model.RoomChat{
		RoomID: 0,
		UserID: me.ID,
	}
	result = db.Create(&room)
	if result.RowsAffected == 0 || result.Error != nil {
		respondError(w, "couldn't prepare insertadfa statement.", http.StatusBadRequest)
		return
	}

	var otherUsers []model.User
	result = db.Where("id != ?", me.ID).Select("id").Find(&otherUsers)
	if result.Error != nil {
		respondError(w, "couldn't prepare insertadfa statement.", http.StatusBadRequest)
		return
	} else if result.RowsAffected == 0 {
		fmt.Println("Zero other users")
		AllowOriginAccess(w)
		w.WriteHeader(http.StatusOK) // 200 OK
		return
	}

	var rooms []model.PrivateChat
	for _, user := range otherUsers {
		rooms = append(rooms, model.PrivateChat{PeerID: user.ID, UserId: me.ID})
	}
	fmt.Println(rooms)
	result = db.Create(rooms)
	if result.Error != nil {
		respondError(w, "couldn't prepare insertadfa statement.", http.StatusBadRequest)
		return
	}

	AllowOriginAccess(w)
	w.WriteHeader(http.StatusOK) // 200 OK
	return
}
