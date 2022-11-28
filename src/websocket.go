package src

import (
	"bytes"
	"chat-backend/model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type wsMessage struct {
	Connect  bool   `json:"connect"`
	Msg      string `json:"message"`
	SenderID int    `json:"senderID"`
	PeerID   int    `json:"peerID"`
	Room     int    `json:"chatroom"`
	DateName string `json:"dateName"`
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	UID int

	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
// func persistToDB(message []byte) {
// 	db, err := model.DbInit()

// }

const NOT_SELECTED = -1

func insertToDB(data *wsMessage) error {
	db, err := model.DbInit()
	if err != nil {
		return err
	}
	now := time.Now()
	fmt.Println(now)
	if data.PeerID == NOT_SELECTED {
		data.DateName = now.Format("2006-01-02 15:04:05")
		fmt.Println("datetime", data.DateName)
		db.Create(&model.RoomChat_t{Message: data.Msg, SenderID: uint(data.SenderID), RoomID: data.Room, Datetime: now})
	} else {
		var privateRoom model.PrivateChat
		res := db.Take(&privateRoom,
			"(user_id = ? AND peer_id = ?) OR (peer_id = ? AND user_id = ?)", data.SenderID, data.PeerID, data.SenderID, data.PeerID)

		if res.Error != nil {
			return res.Error
		}

		data.DateName = now.Format("2006-01-02 15:04:05")
		fmt.Println("datetime", data.DateName)
		db.Create(model.PrivateChat_t{Message: data.Msg, SenderID: uint(data.SenderID), RoomID: privateRoom.RoomID, Datetime: now})
	}
	return nil
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		// from here we start extracting the datas
		var data wsMessage
		json.Unmarshal(message, &data)
		if data.Connect == true {
			c.UID = data.SenderID
			if err != nil {
				fmt.Println("error!")
				c.conn.Close()
			}
			fmt.Println(data)
			fmt.Println("connect request received: ", c.UID, data.SenderID)
		} else {
			fmt.Println("roomID", data.Room)

			if insertToDB(&data) != nil {
				return
			}
			jsonMsg, err := json.Marshal(data)
			if err != nil {
				fmt.Println("error marshaling json")
			}
			c.hub.broadcast <- jsonMsg
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
func ListenWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// r.Header().Set("Origin", r.RemoteAddr)
	if r.Header.Get("Origin") == "http://localhost:3000" ||
		r.Header.Get("Origin") == "www.codeatyolo.link" {
		r.Header.Del("Origin")
		r.Header.Set("Origin", "http://localhost:9999")
	}

	fmt.Println(r.Header.Get("Origin"))
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
