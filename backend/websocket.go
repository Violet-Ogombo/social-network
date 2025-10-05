package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"social-network/backend/db"
	"social-network/backend/models"
	"social-network/backend/utils"

	"github.com/gorilla/websocket"
)

const throttleRate = 500 * time.Millisecond

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients      = make(map[string]*Client)
	clientsMutex sync.RWMutex
)

type Client struct {
	ID       string
	Nickname string
	Conn     *websocket.Conn
	Send     chan []byte
	lastSent time.Time
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(utils.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Get user nickname for the client
	var nickname string
	err = db.DB.QueryRow("SELECT nickname FROM users WHERE id = ?", userID).Scan(&nickname)
	if err != nil {
		log.Println("Error fetching user nickname:", err)
		nickname = userID // fallback
	}

	client := &Client{
		ID:       userID,
		Nickname: nickname,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	clientsMutex.Lock()
	if oldClient, exists := clients[userID]; exists {
		oldClient.Conn.Close() // Triggers cleanup in readPump()
	}
	clients[userID] = client
	clientsMutex.Unlock()

	_, err = db.DB.Exec("UPDATE users SET online_status = 1 WHERE id = ?", userID)
	if err != nil {
		log.Println("Error updating user status:", err)
	}

	sendOnlineUsers("")
	go client.readPump()
	go client.writePump()
}

func (c *Client) readPump() {
	defer func() {
		c.Conn.Close()
		clientsMutex.Lock()
		delete(clients, c.ID)
		clientsMutex.Unlock()
		db.DB.Exec("UPDATE users SET online_status = 0 WHERE id = ?", c.ID)
		db.DB.Exec("UPDATE users SET online_status = 0 WHERE id = ?", c.ID)
		sendOnlineUsers("")
	}()

	for {
		_, msgBytes, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		var msg models.Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Println("Message unmarshal error:", err)
			continue
		}

		// security: always use the authenticated client's ID and nickname as the sender
		msg.SenderID = c.ID
		msg.SenderName = c.Nickname

		switch msg.Type {
		case "message":
			// insert message using authenticated sender id
			result, err := db.DB.Exec(
				`INSERT INTO messages (sender_id, receiver_id, content, created_at)
				VALUES (?, ?, ?, CURRENT_TIMESTAMP)
				`, c.ID, msg.ReceiverID, msg.Content)
			if err != nil {
				log.Println("DB insert error:", err)
				continue
			}
			msgID, _ := result.LastInsertId()
			db.DB.QueryRow("SELECT created_at FROM messages WHERE id = ?", msgID).Scan(&msg.CreatedAt)
			// SenderName already set from authenticated client
			msg.ID = int(msgID)

			encoded, err_ := json.Marshal(msg)
			if err_ != nil {
				log.Println("Message marshal error:", err_)
				continue
			} else {
				fmt.Println("Message marshaled successfully:", string(encoded))
			}

			clientsMutex.RLock()
			receiver, ok := clients[msg.ReceiverID]
			clientsMutex.RUnlock()
			if ok {
				fmt.Println("Sending message to receiver:", receiver.ID)
				receiver.Send <- encoded

				// send a lightweight notification payload
				notification := models.Message{
					Type:       "new_message_notification",
					SenderID:   msg.SenderID,
					SenderName: msg.SenderName,
					Content:    msg.Content,
				}
				notifPayload, _ := json.Marshal(notification)
				receiver.Send <- notifPayload
			} else {
				fmt.Println("Receiver not connected:", msg.ReceiverID)
			}
			c.Send <- encoded

		case "typing":
			// forward typing notification using authenticated sender info
			clientsMutex.RLock()
			receiver, ok := clients[msg.ReceiverID]
			clientsMutex.RUnlock()

			if ok {
				fmt.Println("Forwarding typing notification from", c.ID, "to", receiver.ID)
				typingNotification := models.Message{
					Type:       "typing",
					SenderID:   c.ID,
					SenderName: c.Nickname,
					ReceiverID: msg.ReceiverID,
				}
				payload, _ := json.Marshal(typingNotification)
				receiver.Send <- payload
			}

		case "stop_typing":
			clientsMutex.RLock()
			receiver, ok := clients[msg.ReceiverID]
			clientsMutex.RUnlock()

			if ok {
				stopTypingNotification := models.Message{
					Type:       "stop_typing",
					SenderID:   c.ID,
					ReceiverID: msg.ReceiverID,
				}
				payload, _ := json.Marshal(stopTypingNotification)
				receiver.Send <- payload
			}

		case "user_list_request":
			sendOnlineUsers(c.ID)

		default:
			log.Println("Unknown message type:", msg.Type)
		}
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		// Extract the message type
		var raw struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msg, &raw); err != nil {
			log.Println("Failed to parse message type:", err)
			continue
		}

		// Apply throttling only for typing events
		if raw.Type == "typing" || raw.Type == "stop_typing" {
			if !c.canSendMessage() {
				log.Println("Throttled message:", string(msg))
				continue
			}
		}

		// Send message
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (c *Client) canSendMessage() bool {
	if time.Since(c.lastSent) < throttleRate {
		return false
	}
	c.lastSent = time.Now()
	return true
}

func sendOnlineUsers(_ string) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	for userID, client := range clients {
		go func(userID string, client *Client) {
			rows, err := db.DB.Query(`
SELECT u.id, u.nickname,
	CASE WHEN u.online_status = 1 THEN 1 ELSE 0 END AS is_online,
	MAX(m.created_at) as last_msg
	FROM users u
	LEFT JOIN messages m ON (
		(u.id = m.sender_id AND m.receiver_id = ?) OR
		(u.id = m.receiver_id AND m.sender_id = ?)
	)
	WHERE u.id != ?
	GROUP BY u.id
	ORDER BY
	is_online DESC,                   
	last_msg DESC NULLS LAST,
	u.nickname COLLATE NOCASE ASC
	`, userID, userID, userID)
			if err != nil {
				log.Println("User fetch error:", err)
				return
			}
			defer rows.Close()

			var users []map[string]interface{}
			for rows.Next() {
				var id, nickname string
				var isOnline int
				var lastMsg sql.NullString

				if err := rows.Scan(&id, &nickname, &isOnline, &lastMsg); err != nil {
					continue
				}
				users = append(users, map[string]interface{}{
					"id":        id,
					"nickname":  nickname,
					"is_online": isOnline == 1,
				})
			}

			jsonUsers, _ := json.Marshal(users)
			update := models.Message{Type: "user_list", Content: string(jsonUsers)}
			payload, _ := json.Marshal(update)

			select {
			case client.Send <- payload:
			default:
				close(client.Send)
				client.Conn.Close()
			}
		}(userID, client)
	}
}
