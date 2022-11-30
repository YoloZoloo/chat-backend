// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package src

import (
	"chat-backend/model"
	"encoding/json"
	"fmt"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			// we should implement our own logic here
			var clients []*Client
			var unmarshalMsg wsMessage
			json.Unmarshal(message, &unmarshalMsg)

			if unmarshalMsg.PeerID == NOT_SELECTED {
				var UIDs []int
				db, err := model.DbInit()
				if err != nil {
					fmt.Println("error")
					return
				}
				res := db.Table("room_chats").Select("user_id").Where("room_id = ?", unmarshalMsg.Room).Scan(&UIDs)
				if res.Error != nil {
					fmt.Println("res error")
					return
				}
				for _, uid := range UIDs {
					client, ok := userMap[uid]
					if ok {
						clients = append(clients, client)
					}
				}
			} else {
				client, ok := userMap[unmarshalMsg.SenderID]
				if ok {
					clients = append(clients, client)
				}
				client, ok = userMap[unmarshalMsg.PeerID]
				if ok {
					clients = append(clients, client)
				}
			}
			for _, client := range clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
