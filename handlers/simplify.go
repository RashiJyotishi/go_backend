package handlers

import (
	"go_backend/config"
	"math"
	"sort"

	"github.com/gofiber/fiber/v2"
)

type Tx struct {
	From          int     `json:"from"`
	FromUsername  string  `json:"from_username"`
	To            int     `json:"to"`
	ToUsername    string  `json:"to_username"`
	Amount       float64 `json:"amount"`
}

type Node struct {
	ID       int
	Username string
	Val      float64
}

func SimplifyGroup(c *fiber.Ctx) error {

	groupID := c.Params("id")

	rows, err := config.DB.Query(`
	WITH paid AS (
		SELECT payer_id AS user_id, SUM(amount) AS p
		FROM expenses
		WHERE group_id=$1
		GROUP BY payer_id
	),
	owed AS (
		SELECT s.user_id, SUM(s.amount_owed) AS o
		FROM splits s
		JOIN expenses e ON e.id=s.expense_id
		WHERE e.group_id=$1
		GROUP BY s.user_id
	)
	SELECT gm.user_id,u.username,
	       COALESCE(p.p,0) - COALESCE(o.o,0) AS net
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		LEFT JOIN paid p ON p.user_id=gm.user_id
		LEFT JOIN owed o ON o.user_id=gm.user_id
		WHERE gm.group_id=$1
	`, groupID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var debtors []Node
	var creditors []Node

	for rows.Next() {
		var id int
		var net float64
		var username string

		if err := rows.Scan(&id, &username, &net); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		if net < 0 {
			debtors = append(debtors, Node{ID: id, Username: username, Val: -net})
		} else if net > 0 {
			creditors = append(creditors, Node{ID: id, Username: username, Val: net})
		}
	}

	if err := rows.Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Sort largest first (Splitwise logic)
	sort.Slice(debtors, func(i, j int) bool {
		return debtors[i].Val > debtors[j].Val
	})

	sort.Slice(creditors, func(i, j int) bool {
		return creditors[i].Val > creditors[j].Val
	})

	i := 0
	j := 0

	var result []Tx

	for i < len(debtors) && j < len(creditors) {

		d := &debtors[i]
		cn := &creditors[j]

		amt := math.Min(d.Val, cn.Val)

		amt = round2(amt)

		result = append(result, Tx{
			From:        d.ID,
			FromUsername: d.Username,
			To:          cn.ID,
			ToUsername:  cn.Username,
			Amount:      amt,
		})

		d.Val -= amt
		cn.Val -= amt

		if d.Val < 0.01 {
			i++
		}

		if cn.Val < 0.01 {
			j++
		}
	}

	return c.JSON(result)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
