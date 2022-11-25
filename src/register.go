package src

import (
	model "chat-backend/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type registerForm struct {
	UserID    string `json:"user_id"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
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

	db, err := model.DbInit()
	if err != nil {
		respondError(w, "", http.StatusInternalServerError)
		return
	}

	var cnt int
	adminAuthority := 0
	result := db.Raw("SELECT count(*) FROM users").Scan(&cnt)
	if result.Error != nil && cnt > 0 {
		adminAuthority = 1
	}

	hashedPassword, _ := HashPassword(data.Password)
	user := model.User{
		Username:       data.UserID,
		Password:       hashedPassword,
		Firstname:      data.FirstName,
		Lastname:       data.LastName,
		AdminAuthority: adminAuthority,
	}
	result = db.Create(&user)
	if result.RowsAffected == 0 || result.Error != nil {
		respondError(w, "couldn't prepare insertadfa statement.", http.StatusBadRequest)
		return
	}

	var me model.User
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
}

