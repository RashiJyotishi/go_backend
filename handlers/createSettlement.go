package handlers

import (
	"go_backend/config"
	"github.com/gofiber/fiber/v2"
)

type SettleReq struct {
    GroupID int     `json:"group_id"`
    PayeeID int     `json:"payee_id"`
    Amount  float64 `json:"amount"`
}


func CreateSettlement(c *fiber.Ctx) error {
	uid := c.Locals("user_id")
	if uid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	payerID := uid.(int)

	req := new(SettleReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad Request"})
	}
	_, err := config.DB.Exec(`
		INSERT INTO settlements (group_id, payer_id, payee_id, amount)
        VALUES ($1, $2, $3, $4)`,
        req.GroupID, payerID, req.PayeeID, req.Amount)

	if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
	return c.Status(201).JSON(fiber.Map{"message": "Settlement recorded"})
}
