package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SubscribedRooms struct {
	GroupRoom   []string   `json:"grouproom_id"`
	PrivateChat []PeerInfo `json:"privateroom_id"`
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

	db, err := OpenDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	queryString := "SELECT grouproom_id FROM chat.grouproom_m WHERE guest_id = ?"
	rows, err := db.Query(queryString, uniqueID)
	if err != nil {
		fmt.Println("here")
		http.Error(w, err.Error(), http.StatusNonAuthoritativeInfo)
		return
	}
	defer rows.Close()

	var groupRooms []string
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var room string
		if err := rows.Scan(&room); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		groupRooms = append(groupRooms, room)
	}
	rows.Close()

	queryString = `SELECT user.user_id, user.id as id FROM chat.user_m as user
					RIGHT OUTER JOIN (SELECT idA from chat.privateroom_m WHERE idB = ? ) as pr
        			ON user.id = pr.idA`

	rows, err = db.Query(queryString, uniqueID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNonAuthoritativeInfo)
		return
	}
	defer rows.Close()

	var privateRooms []PeerInfo
	for rows.Next() {
		var peer_userID string
		var peer_id string

		if err := rows.Scan(&peer_userID, &peer_id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(privateRooms)
		privateRooms = append(privateRooms, PeerInfo{ID: peer_id, UserID: peer_userID})
		// privateRooms = append(privateRooms, peer_userID)
	}
	rows.Close()

	queryString = `SELECT user.user_id, user.id as id
					FROM chat.user_m as user
					RIGHT OUTER JOIN (SELECT idB from chat.privateroom_m WHERE idA = ? ) as pr
					ON user.id = pr.idB`

	rows, err = db.Query(queryString, uniqueID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNonAuthoritativeInfo)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var peer_userID string
		var peer_id string

		if err := rows.Scan(&peer_userID, &peer_id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		privateRooms = append(privateRooms, PeerInfo{ID: peer_id, UserID: peer_userID})
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
	return
}
