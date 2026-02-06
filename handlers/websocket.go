package handlers

import (
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type Hub struct {
	Clients map[*websocket.Conn]bool
	Mutex   sync.Mutex
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

	log.Printf("client join this group %s", groupID)

	defer func() {
		currentHub.Mutex.Lock()
		delete(currentHub.Clients, c)
		currentHub.Mutex.Unlock()

		c.Close()
		log.Printf("Client left Group %s", groupID)
	}()


	for {
		mt, msg, err := c.ReadMessage()
		log.Printf("Message received in Room: %s", groupID)
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		currentHub.Mutex.Lock()
		clientCount := len(currentHub.Clients)
		currentHub.Mutex.Unlock()
		log.Printf("Broadcasting to %d clients in Room %s", clientCount, groupID)

		log.Printf("Group %s message: %s", groupID, msg)

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