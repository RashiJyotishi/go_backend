package handlers

import (
	"encoding/json"
	"go_backend/config"
	"log"
	"strconv"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type Hub struct {
	Clients map[*websocket.Conn]bool
	Mutex   sync.Mutex
}

type IncomingMessage struct {
    Type     string `json:"type"`
    Message  string `json:"message"`
    FileURL  string `json:"file_url"`
    FileType string `json:"file_type"`
    UserID   string `json:"userID"`
	MessageID int    `json:"messageID"`
}

var GlobalHub = Hub{
	Clients: make(map[*websocket.Conn]bool),
}

var (
	Rooms      = make(map[string]*Hub)
	RoomsMutex sync.Mutex
)

func getHub(groupID string) *Hub {
	RoomsMutex.Lock()
	defer RoomsMutex.Unlock()

	hub, exists := Rooms[groupID]
	if !exists {
		hub = &Hub{
			Clients: make(map[*websocket.Conn]bool),
		}
		Rooms[groupID] = hub
		log.Printf("Created new room for Group %s", groupID)
	}
	return hub
}

func WebsocketHandler(c *websocket.Conn) {
	groupID := c.Params("id")

	if groupID == "" {
		log.Println("No group ID provided")
		c.Close()
		return
	}

	currentHub := getHub(groupID)
	currentHub.Mutex.Lock()
	currentHub.Clients[c] = true
	currentHub.Mutex.Unlock()
	var userIDInt int

	log.Printf("client join this group %s", groupID)

	defer func() {
		currentHub.Mutex.Lock()
		delete(currentHub.Clients, c)
		currentHub.Mutex.Unlock()

		c.Close()
		log.Printf("Client left Group %s", groupID)
		_, err := config.DB.Exec("UPDATE users SET is_online = FALSE WHERE id = $1", userIDInt)
		if err != nil {
			log.Println("Error setting user offline:", err)
		}
	}()


	for {
		mt, msg, err := c.ReadMessage()
		log.Printf("Message received in Room: %s", groupID)
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		var parsedMsg IncomingMessage
		if err := json.Unmarshal(msg, &parsedMsg); err != nil {
			log.Println("Error parsing JSON:", err)
			break
		}

		var fileTypeToSave interface{}
		if parsedMsg.FileType == "" {
			fileTypeToSave = nil
		} else {
			fileTypeToSave = parsedMsg.FileType
		}

		var groupIDInt int
		var messageID int

		userIDInt, err = strconv.Atoi(parsedMsg.UserID)
		if err != nil {
			log.Println("Invalid User ID:", err)
			break
		}

		groupIDInt, err = strconv.Atoi(groupID)
		if err != nil {
			log.Println("Invalid Group ID:", err)
			break
		}

		var exists int

		err = config.DB.QueryRow(`
			SELECT 1
			FROM group_members
			WHERE group_id = $1 AND user_id = $2
			LIMIT 1`,
			groupIDInt, userIDInt).Scan(&exists)

		if err != nil {

			log.Printf("⚠️ SECURITY ALERT: User %d attempted to spoof Group %s", userIDInt, groupID)
			broadcastMsg := IncomingMessage{
				Type:      "error",
				Message:   "Access Denied: You are not a member of this group.",
				UserID:    parsedMsg.UserID,
				MessageID: 0,
			}

			parsedBroadcastMsg, _ := json.Marshal(broadcastMsg)

			currentHub.Mutex.Lock()
			c.WriteMessage(mt, parsedBroadcastMsg)
			currentHub.Mutex.Unlock()

			break
		}
		_, err = config.DB.Exec("UPDATE users SET is_online = TRUE WHERE id = $1", userIDInt)
		if err != nil {
			log.Println("Error setting user online:", err)
		}

		log.Printf("Group %s message: %s", groupID, msg)

		// currentHub.Mutex.Lock()
		// clientCount := len(currentHub.Clients)
		// currentHub.Mutex.Unlock()
		// log.Printf("Broadcasting to %d clients in Room %s", clientCount, groupID)

		log.Printf("Group %s message: %s", groupID, msg)

		switch parsedMsg.Type {
		case "chat":
			log.Println("Chat message received")
			err = config.DB.QueryRow(`
				INSERT INTO messages (user_id, group_id, message, file_url, file_type)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id
				`,
				userIDInt,
				groupIDInt,
				parsedMsg.Message,
				parsedMsg.FileURL,
				fileTypeToSave,
				).Scan(&messageID)

			if err != nil {
				log.Println("Error saving message to DB:", err)
				break
			}
			broadcastMsg := IncomingMessage{
				Type:      parsedMsg.Type,
				Message:   parsedMsg.Message,
				FileURL:   parsedMsg.FileURL,
				FileType:  parsedMsg.FileType,
				UserID:    parsedMsg.UserID,
				MessageID: messageID,
			}

			parsedBroadcastMsg, err := json.Marshal(broadcastMsg)
			if err != nil {
				log.Println("Error marshaling broadcast message:", err)
				break
			}
			log.Println("Broadcasting message to clients in Group", groupID)

			currentHub.Mutex.Lock()
			for client := range currentHub.Clients {

				if err := client.WriteMessage(mt, parsedBroadcastMsg); err != nil {
					log.Println("Write error:", err)
					client.Close()
					delete(currentHub.Clients, client)
				}
			}
			currentHub.Mutex.Unlock()
		case "status":
			log.Println("Status update received")

			currentHub.Mutex.Lock()
			for client := range currentHub.Clients {
				if err := client.WriteMessage(mt, msg); err != nil {
					log.Println("Write error:", err)
					client.Close()
					delete(currentHub.Clients, client)
				}
			}
			currentHub.Mutex.Unlock()
		}

	}
}