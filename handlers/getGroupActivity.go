package handlers

import (
	"go_backend/config"
	"github.com/gofiber/fiber/v2"
    "time"
)

type FeedItem struct {
    PayerID     int     `json:"payer_id"`
    Amount      float64 `json:"amount"`
    Description string  `json:"description"`
    CreatedAt   string  `json:"created_at"` // or time.Time
    PayeeID     int     `json:"payee_id"`   // Will be 0 for expenses
}

type ChatMessage struct {
    UserID    int       `json:"user_id"`
    Username  string    `json:"username"`
    Message   string    `json:"message"`
    FileURL   *string   `json:"file_url"`
    FileType  *string   `json:"file_type"`
    CreatedAt time.Time `json:"created_at"`
}

func GetGroupActivity(c *fiber.Ctx) error {
    groupID := c.Params("id")

    chatQuery := `
        SELECT m.user_id, u.username, m.message, m.file_url, m.file_type, m.created_at
        FROM messages m
        JOIN users u ON m.user_id = u.id
        WHERE m.group_id = $1
        ORDER BY m.created_at ASC`

    rows, _ := config.DB.Query(chatQuery, groupID)
    defer rows.Close()

    var history []ChatMessage
    for rows.Next() {
        var msg ChatMessage
        rows.Scan(&msg.UserID, &msg.Username, &msg.Message, &msg.FileURL, &msg.FileType, &msg.CreatedAt)
        history = append(history, msg)
    }

    // Your UNION query
    query := `
    (SELECT payer_id, amount, description, created_at, 0 as payee_id 
     FROM expenses 
     WHERE group_id = $1)
    UNION ALL
    (SELECT payer_id, amount, 'Settlement', created_at, payee_id 
     FROM settlements 
     WHERE group_id = $1)
    ORDER BY created_at DESC
    `

    rows, err := config.DB.Query(query, groupID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    defer rows.Close()

    var feed []FeedItem

    // Loop through the mixed results
    for rows.Next() {
        var item FeedItem
        // Scan must match the order of columns in the SELECT
        if err := rows.Scan(&item.PayerID, &item.Amount, &item.Description, &item.CreatedAt, &item.PayeeID); err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Error scanning data"})
        }
        feed = append(feed, item)
    }

    return c.JSON(fiber.Map{
        "activity_feed": feed,    // your existing financial data
        "chat_history":  history, // the new chat history
    })
}