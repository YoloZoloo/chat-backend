package src

import (
	"chat-backend/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ChatFromDb struct {
	MessageID uint
	Message   string
	DateTime  time.Time
	Lastname  string
	Firstname string
	SenderID  string
}

type ChatMessages struct {
	MessageID uint   `json:"messageID"`
	Message   string `json:"message"`
	DateName  string `json:"dateName"`
	SenderID  string `json:"senderID"`
}
type RoomID struct {
	Room string `json:"room"`
}

type PeerID struct {
	PeerID string `json:"peer"`
}

func GetGroupChat(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	uniqueID, err := ValidateToken(token)

	if err != nil {
		respondError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	reqBody, _ := io.ReadAll(r.Body)
	var data RoomID
	json.Unmarshal(reqBody, &data)
	fmt.Println("roomID", data.Room)

	db, err := model.DbInit()
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var chat []ChatFromDb
	res := db.Raw(`SELECT chat.message_id, chat.message, chat.datetime,
	IFNULL(user.lastname, "") as lastname, IFNULL(user.firstname, "") as firstname, IFNULL(user.id, "") as id FROM
	(SELECT chatR.message_id, chatR.message, chatR.datetime, chatR.sender_id
		FROM chat.room_chat_ts as chatR
		INNER JOIN
		(SELECT * FROM chat.room_chats WHERE guest_id = ?) as gm 
			ON gm.room_id = chatR.room_id and gm.room_id = ?
			ORDER BY message_id desc LIMIT 0 , 20) as chat
	LEFT OUTER JOIN chat.users as user ON user.id = chat.sender_id
	ORDER BY chat.message_id ASC`, uniqueID, data.Room).Scan(&chat)

	if res.Error != nil {
		respondError(w, res.Error.Error(), http.StatusInternalServerError)
		return
	}

	var messages []ChatMessages
	// Loop through rows, using Scan to assign column data to struct fields.
	for _, row := range chat {
		messages = append(messages, ChatMessages{MessageID: row.MessageID,
			Message:  row.Message,
			DateName: row.DateTime.String() + ": " + row.Firstname + " " + row.Lastname,
			SenderID: row.SenderID})
	}

	respData, err := json.Marshal(messages)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	AllowOriginAccess(w)
	w.Write(respData)
}

func GetPrivateChat(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	uniqueID, err := ValidateToken(token)
	if err != nil {
		respondError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	reqBody, _ := io.ReadAll(r.Body)
	var data PeerID
	json.Unmarshal(reqBody, &data)
	fmt.Println("peerID", data.PeerID)

	db, err := model.DbInit()
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var roomID int
	res := db.Find(&model.PrivateChat{}, "(user_id = ? AND peer_id = ?) OR (peer_id = ? AND user_id = ?)").Scan(&roomID)

	if res.Error != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("privRoomID", roomID)

	var chat []ChatFromDb
	res = db.Raw(`SELECT chat.message_id, chat.message, chat.datetime,
		user.lastname, user.firstname, user.id 
		FROM
		(SELECT chat_t.message_id, chat_t.message, chat_t.datetime, chat_t.sender_id
			FROM chat.private_chat_ts as chat_t
		INNER JOIN
		(SELECT * FROM chat.private_chats WHERE peer_id = ? or user_id = ?) as pm 
		ON pm.room_id = chat_t.room_id and pm.room_id = ?
		ORDER BY message_id desc LIMIT 0 , 20) as chat
		LEFT OUTER JOIN chat.users as user ON user.id = chat.sender_id
		ORDER BY chat.message_id ASC`, uniqueID, uniqueID, roomID).
		Scan(&chat)

	if res.Error != nil {
		respondError(w, res.Error.Error(), http.StatusNonAuthoritativeInfo)
		return
	}

	var messages []ChatMessages
	// Loop through rows, using Scan to assign column data to struct fields.
	for _, row := range chat {
		messages = append(messages, ChatMessages{MessageID: row.MessageID,
			Message:  row.Message,
			DateName: row.DateTime.String() + ": " + row.Firstname + " " + row.Lastname,
			SenderID: row.SenderID})
	}

	respData, err := json.Marshal(messages)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	AllowOriginAccess(w)
	w.Write(respData)
}

