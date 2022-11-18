package src

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ChatMessages struct {
	MessageID int16  `json:"messageID"`
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

	reqBody, _ := ioutil.ReadAll(r.Body)
	var data RoomID
	json.Unmarshal(reqBody, &data)
	roomID := data.Room
	fmt.Println("roomID", data.Room)

	db, err := OpenDB()
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	queryString :=
		`SELECT chat.message_id, chat.message, chat.datetime,
			IFNULL(user.lastname, "") as lastname, IFNULL(user.firstname, "") as firstname, IFNULL(user.id, "") as id FROM
			(SELECT chatR.message_id, chatR.message, chatR.datetime, chatR.sender_id
				FROM chat.groupchat_t as chatR
				INNER JOIN
				(SELECT * FROM chat.grouproom_m WHERE guest_id = ?) as gm 
					ON gm.grouproom_id = chatR.grouproom_id and gm.grouproom_id = ?
					ORDER BY message_id desc LIMIT 0 , 20) as chat
			LEFT OUTER JOIN chat.user_m as user ON user.id = chat.sender_id
			ORDER BY chat.message_id ASC`

	rows, err := db.Query(queryString, uniqueID, roomID)

	if err != nil {
		respondError(w, err.Error(), http.StatusNonAuthoritativeInfo)
		return
	}

	defer rows.Close()

	var messages []ChatMessages
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var MessageID int16
		var Message string
		var DateTime string
		var LastName string
		var FirstName string
		var SenderID string
		if err := rows.Scan(&MessageID, &Message, &DateTime, &LastName, &FirstName, &SenderID); err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		messages = append(messages, ChatMessages{MessageID: MessageID,
			Message:  Message,
			DateName: DateTime + ": " + FirstName + " " + LastName,
			SenderID: SenderID})
	}
	rows.Close()
	respData, err := json.Marshal(messages)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	AllowOriginAccess(w)
	w.Write(respData)
	return
}
func GetPrivateChat(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	uniqueID, err := ValidateToken(token)
	if err != nil {
		respondError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	reqBody, _ := ioutil.ReadAll(r.Body)
	var data PeerID
	json.Unmarshal(reqBody, &data)

	peer := data.PeerID
	fmt.Println("peerID", peer)

	db, err := OpenDB()
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roomID, err := GetPrivRoomID(uniqueID, peer)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("privRoomID", roomID)

	queryString :=
		`SELECT chat.message_id, chat.message, chat.datetime,
			user.lastname, user.firstname, user.id 
			FROM
			(SELECT chat_t.message_id, chat_t.message, chat_t.datetime, chat_t.sender_id
				FROM chat.privatechat_t as chat_t
			INNER JOIN
			(SELECT * FROM chat.privateroom_m WHERE idA = ? or idB = ?) as pm 
			ON pm.privateroom_id = chat_t.privateroom_id and pm.privateroom_id = ?
			ORDER BY message_id desc LIMIT 0 , 20) as chat
			LEFT OUTER JOIN chat.user_m as user ON user.id = chat.sender_id
			ORDER BY chat.message_id ASC `

	rows, err := db.Query(queryString, uniqueID, uniqueID, roomID)

	if err != nil {
		respondError(w, err.Error(), http.StatusNonAuthoritativeInfo)
		return
	}

	defer rows.Close()

	var messages []ChatMessages
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var MessageID int16
		var Message string
		var DateTime string
		var LastName string
		var FirstName string
		var SenderID string
		if err := rows.Scan(&MessageID, &Message, &DateTime, &LastName, &FirstName, &SenderID); err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		messages = append(messages, ChatMessages{MessageID: MessageID,
			Message:  Message,
			DateName: DateTime + ": " + FirstName + " " + LastName,
			SenderID: SenderID})
	}
	rows.Close()
	respData, err := json.Marshal(messages)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	AllowOriginAccess(w)
	w.Write(respData)
	return
}

func GetPrivRoomID(uniqID int16, peerID string) (int, error) {
	db, err := OpenDB()
	if err != nil {
		return 0, err
	}

	queryString := `SELECT privateroom_id from privateroom_m 
		WHERE (idA = ? AND idB = ?) OR (idB = ? AND idA = ?)`
	rows, err := db.Query(queryString, uniqID, peerID, uniqID, peerID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var privateRoomID int
	for rows.Next() {
		if err := rows.Scan(&privateRoomID); err != nil {
			return 0, err
		}
	}
	rows.Close()
	return privateRoomID, err
}
