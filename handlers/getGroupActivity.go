package handlers

import (
	"go_backend/config"
	"github.com/gofiber/fiber/v2"
)

type FeedItem struct {
    PayerID     int     `json:"payer_id"`
    Amount      float64 `json:"amount"`
    Description string  `json:"description"`
    CreatedAt   string  `json:"created_at"` // or time.Time
    PayeeID     int     `json:"payee_id"`   // Will be 0 for expenses
}

func GetGroupActivity(c *fiber.Ctx) error {
    groupID := c.Params("id")

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

    return c.JSON(feed)
}