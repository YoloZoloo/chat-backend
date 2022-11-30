package src

import (
	"chat-backend/model"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type ChatFromDb struct {
	MessageID uint
	Message   string
	Datetime  string
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
	PeerID int `json:"peer_id"`
}

var layout = "2006-01-02 15:04:05"

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

	db, err := model.DbInit()
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var chat []ChatFromDb
	res := db.Raw(`SELECT chats.message_id, chats.message, chats.datetime,
	IFNULL(user.lastname, "") as lastname, IFNULL(user.firstname, "") as firstname, IFNULL(user.id, "") as id FROM
	(SELECT chatR.message_id, chatR.message, chatR.datetime, chatR.sender_id
		FROM chat.room_chat_ts as chatR
		INNER JOIN
		(SELECT * FROM chat.room_chats WHERE user_id = ?) as gm 
			ON gm.room_id = chatR.room_id and gm.room_id = ?
			ORDER BY message_id desc LIMIT 0 , 20) as chats
	LEFT OUTER JOIN chat.users as user ON user.id = chats.sender_id
	ORDER BY chats.message_id ASC`, uniqueID, data.Room).Scan(&chat)

	if res.Error != nil {
		respondError(w, res.Error.Error(), http.StatusInternalServerError)
		return
	}

	var messages []ChatMessages
	// Loop through rows, using Scan to assign column data to struct fields.
	if res.RowsAffected == 0 {
		messages = []ChatMessages{}
	} else {
		for _, row := range chat {
			datetime, err := time.Parse(layout, row.Datetime)
			if err != nil {
				messages = append(messages, ChatMessages{MessageID: row.MessageID,
					Message:  row.Message,
					DateName: row.Datetime + ": " + row.Firstname + " " + row.Lastname,
					SenderID: row.SenderID})
			} else {
				messages = append(messages, ChatMessages{MessageID: row.MessageID,
					Message:  row.Message,
					DateName: datetime.Format(layout) + ": " + row.Firstname + " " + row.Lastname,
					SenderID: row.SenderID})
			}

		}
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

	db, err := model.DbInit()
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var privateRoom model.PrivateChat
	res := db.Take(&privateRoom,
		"(user_id = ? AND peer_id = ?) OR (peer_id = ? AND user_id = ?)", uniqueID, data.PeerID, uniqueID, data.PeerID)

	if res.Error != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var chat []ChatFromDb
	res = db.Raw(`SELECT chats.message_id, chats.message, chats.datetime,
		user.lastname, user.firstname, user.id 
		FROM
		(SELECT chat_t.message_id, chat_t.message, chat_t.datetime, chat_t.sender_id
			FROM chat.private_chat_ts as chat_t
		INNER JOIN
		(SELECT * FROM chat.private_chats WHERE peer_id = ? or user_id = ?) as pm 
		ON pm.room_id = chat_t.room_id and pm.room_id = ?
		ORDER BY message_id desc LIMIT 0 , 20) as chats
		LEFT OUTER JOIN chat.users as user ON user.id = chats.sender_id
		ORDER BY chats.message_id ASC`, uniqueID, uniqueID, privateRoom.RoomID).
		Scan(&chat)

	if res.Error != nil {
		respondError(w, res.Error.Error(), http.StatusNonAuthoritativeInfo)
		return
	}

	var messages []ChatMessages
	// Loop through rows, using Scan to assign column data to struct fields.
	if res.RowsAffected == 0 {
		messages = []ChatMessages{}
	} else {
		for _, row := range chat {
			datetime, err := time.Parse(layout, row.Datetime)
			if err != nil {
				messages = append(messages, ChatMessages{MessageID: row.MessageID,
					Message:  row.Message,
					DateName: row.Datetime + ": " + row.Firstname + " " + row.Lastname,
					SenderID: row.SenderID})
			} else {
				messages = append(messages, ChatMessages{MessageID: row.MessageID,
					Message:  row.Message,
					DateName: datetime.Format(layout) + ": " + row.Firstname + " " + row.Lastname,
					SenderID: row.SenderID})
			}

		}
	}

	respData, err := json.Marshal(messages)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	AllowOriginAccess(w)
	w.Write(respData)
}
