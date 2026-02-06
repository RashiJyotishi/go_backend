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

// 3. The Handler
// In Fiber, the handler signature for websockets is just func(*websocket.Conn)
func WebsocketHandler(c *websocket.Conn) {
	// groupID := c.Params("id")
	// A. Add to Hub
	GlobalHub.Mutex.Lock()
	GlobalHub.Clients[c] = true
	GlobalHub.Mutex.Unlock()

	log.Println("New client connected!")

	defer func() {
		// B. Cleanup logic
		GlobalHub.Mutex.Lock()
		delete(GlobalHub.Clients, c)
		GlobalHub.Mutex.Unlock()
		c.Close()
		log.Println("Client disconnected.")
	}()

	// 4. The Infinite Loop ♾️
	for {
		// Read Message
		// mt is message type (Text or Binary), msg is the actual data
		mt, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		log.Printf("Received: %s", msg)

		// 5. Broadcast (Send to everyone)
		GlobalHub.Mutex.Lock()
		for client := range GlobalHub.Clients {
			// Write the message back to the client
			if err := client.WriteMessage(mt, msg); err != nil {
				log.Println("Write error:", err)
				client.Close()
				delete(GlobalHub.Clients, client)
			}
		}
		GlobalHub.Mutex.Unlock()
	}
}