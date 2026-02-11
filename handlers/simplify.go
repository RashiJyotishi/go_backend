package handlers

import (
	"go_backend/config"
	"math"
	"sort"

	"github.com/gofiber/fiber/v2"
)

type Tx struct {
	From         int     `json:"from"`
	FromUsername string  `json:"from_username"`
	To           int     `json:"to"`
	ToUsername   string  `json:"to_username"`
	Amount       float64 `json:"amount"`
}

type Node struct {
	ID       int
	Username string
	Val      float64
}

func SimplifyGroup(c *fiber.Ctx) error {
	groupID := c.Params("id")
	if groupID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Group ID required"})
	}

	// 1. Calculate Net Balance for every user
	// Net = (Money Put In) - (Money Consumed)
	// Money Put In = Expenses Paid + Settlements Paid
	// Money Consumed = Expense Splits Owed + Settlements Received
	query := `
	WITH
	expense_paid AS (
		SELECT payer_id AS user_id, SUM(amount)::FLOAT as amt
		FROM expenses WHERE group_id=$1 GROUP BY payer_id
	),
	expense_share AS (
		SELECT s.user_id, SUM(s.amount_owed)::FLOAT as amt
		FROM splits s
		JOIN expenses e ON s.expense_id = e.id
		WHERE e.group_id=$1 GROUP BY s.user_id
	),
	settlement_paid AS (
		SELECT payer_id AS user_id, SUM(amount)::FLOAT as amt
		FROM settlements WHERE group_id=$1 GROUP BY payer_id
	),
	settlement_received AS (
		SELECT payee_id AS user_id, SUM(amount)::FLOAT as amt
		FROM settlements WHERE group_id=$1 GROUP BY payee_id
	)
	SELECT
		gm.user_id,
		u.username,
		(
			COALESCE(ep.amt, 0) + COALESCE(sp.amt, 0)
			-
			(COALESCE(es.amt, 0) + COALESCE(sr.amt, 0))
		) AS net_balance
	FROM group_members gm
	JOIN users u ON gm.user_id = u.id
	LEFT JOIN expense_paid ep ON gm.user_id = ep.user_id
	LEFT JOIN expense_share es ON gm.user_id = es.user_id
	LEFT JOIN settlement_paid sp ON gm.user_id = sp.user_id
	LEFT JOIN settlement_received sr ON gm.user_id = sr.user_id
	WHERE gm.group_id=$1
	`

	rows, err := config.DB.Query(query, groupID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error: " + err.Error()})
	}
	defer rows.Close()

	var debtors []Node
	var creditors []Node

	// 2. Separate users into Debtors and Creditors
	for rows.Next() {
		var id int
		var username string
		var net float64

		if err := rows.Scan(&id, &username, &net); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Scan error: " + err.Error()})
		}

		net = round2(net)

		if net < -0.01 {
			debtors = append(debtors, Node{ID: id, Username: username, Val: -net}) // Store as positive magnitude
		} else if net > 0.01 {
			creditors = append(creditors, Node{ID: id, Username: username, Val: net})
		}
	}

	if err := rows.Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. Greedy Matching Algorithm
	// Sort by magnitude (largest amounts first) to minimize number of transactions
	sort.Slice(debtors, func(i, j int) bool {
		return debtors[i].Val > debtors[j].Val
	})
	sort.Slice(creditors, func(i, j int) bool {
		return creditors[i].Val > creditors[j].Val
	})

	var result []Tx
	i := 0
	j := 0

	// Match debtors to creditors
	for i < len(debtors) && j < len(creditors) {
		debtor := &debtors[i]
		creditor := &creditors[j]

		// The amount to settle is the minimum of what's owed vs what's receivable
		amount := math.Min(debtor.Val, creditor.Val)
		amount = round2(amount)

		// Record the transaction
		if amount > 0 {
			result = append(result, Tx{
				From:         debtor.ID,
				FromUsername: debtor.Username,
				To:           creditor.ID,
				ToUsername:   creditor.Username,
				Amount:       amount,
			})
		}

		// Adjust remaining balances
		debtor.Val -= amount
		creditor.Val -= amount

		// Move to next person if their balance is settled (close to 0)
		if debtor.Val < 0.01 {
			i++
		}
		if creditor.Val < 0.01 {
			j++
		}
	}

	// 4. Return the list of suggested transactions
	return c.JSON(result)
}

// Helper to round to 2 decimal places to avoid floating point weirdness
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}