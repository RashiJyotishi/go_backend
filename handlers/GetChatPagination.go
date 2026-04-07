package handlers

import (
	"go_backend/config"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetChatPagination(c *fiber.Ctx) error {
	// log.Println("Received request for chat pagination")
	groupID := c.Params("id")
	beforeID := c.Query("before_id") // The "Cursor"
	limit := 20

	// 1. Build Query: Fetch messages older than beforeID
	query := `
		SELECT m.user_id, u.username, m.message, m.file_url, m.file_type, m.id
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.group_id = $1`
	
	args := []interface{}{groupID}

	if beforeID != "" {
		query += " AND m.id < $2"
		args = append(args, beforeID)
	}

	// We use DESC to get the 20 messages immediately preceding the cursor
	query += " ORDER BY m.id DESC LIMIT $3"
	args = append(args, limit)
	// log.Print(query);
	rows, err := config.DB.Query(query, args...)
	// log.Print(rows);
	if err != nil {
		// log.Println("Pagination Query Error:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch history"})
	}
	defer rows.Close()

	var messages []IncomingMessage
	for rows.Next() {
		var msg IncomingMessage
		var tempUserID int
		if err := rows.Scan(&tempUserID, &msg.Username, &msg.Message, &msg.FileURL, &msg.FileType, &msg.MessageID); err == nil {
			msg.UserID = strconv.Itoa(tempUserID)
			msg.Type = "chat"
			messages = append(messages, msg)
			// log.Printf("Fetched message ID %d for user %s", msg.MessageID, msg.Username)
		}
	}

	// 2. Reverse: Flip the slice so it's [Oldest -> Newest] for the frontend
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	// log.Printf("Fetched %d messages for group %s with before_id %s", len(messages), groupID, beforeID)
	return c.JSON(fiber.Map{
		"messages": messages,
		"has_more": len(messages) == limit,
	})
}