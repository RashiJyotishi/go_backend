package handlers

import (
	"go_backend/config"
	"time"
	"github.com/gofiber/fiber/v2"
)

func GetGroupExpenses(c *fiber.Ctx) error {
    // Get the group ID from the URL
    groupID := c.Params("id")
    if groupID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Group ID is required",
        })
    }

	rows, _ := config.DB.Query(`SELECT
		e.amount, e.description, e.created_at, u.username
		FROM expenses e
		JOIN users u ON e.payer_id = u.id WHERE e.group_id=$1;`, groupID);
    var groups []fiber.Map

	for rows.Next() {
		var amount float64
		var description string
		var created_at time.Time
		var username string

		rows.Scan(&amount, &description, &created_at, &username);
		groups = append(groups, fiber.Map{
			"amount": amount,
			"description": description,
			"created_at": created_at,
			"username": username,
		})
	}
	if err := rows.Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(groups)
}