package src

import (
	"chat-backend/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type SubscribedRooms struct {
	GroupRoom   []string   `json:"grouproom_id"`
	PrivateChat []PeerInfo `json:"privateroom_id"`
}
type PeerInfoDb struct {
	Username string
	ID       int
}
type PeerInfo struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

func GetSubscribedRooms(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	uniqueID, err := ValidateToken(token)
	if err != nil {
		AllowOriginAccess(w)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	db, err := model.DbInit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var roomChat []model.RoomChat
	res := db.Find(&roomChat, "user_id = ?", uniqueID)
	if res.Error != nil || res.RowsAffected == 0 {
		http.Error(w, err.Error(), http.StatusNonAuthoritativeInfo)
		return
	}

	var groupRooms []string
	// Loop through rows, using Scan to assign column data to struct fields.

	if res.RowsAffected == 0 {
		groupRooms = []string{}
	} else {
		for _, row := range roomChat {
			groupRooms = append(groupRooms, strconv.Itoa(row.RoomID))
		}
	}

	var peers []PeerInfoDb
	res = db.Raw(`SELECT user.username, user.id FROM chat.users as user
		RIGHT OUTER JOIN (
			SELECT user_id as id from chat.private_chats WHERE peer_id = ? 
			UNION ALL 
			SELECT peer_id as id from chat.private_chats WHERE user_id = ?
		) as pr
		ON user.id = pr.id`, uniqueID, uniqueID).Scan(&peers)

	if res.Error != nil {
		http.Error(w, err.Error(), http.StatusNonAuthoritativeInfo)
		return
	}

	var privateRooms []PeerInfo
	if res.RowsAffected == 0 {
		privateRooms = []PeerInfo{}
	} else {
		for _, row := range peers {
			privateRooms = append(privateRooms, PeerInfo{ID: strconv.Itoa(row.ID), UserID: row.Username})
		}
	}

	respData, err := json.Marshal(
		&SubscribedRooms{GroupRoom: groupRooms, PrivateChat: privateRooms})
	fmt.Println(string(respData))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	AllowOriginAccess(w)
	w.Write(respData)
}

