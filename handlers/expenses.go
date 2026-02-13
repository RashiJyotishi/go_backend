package handlers

import (
	"go_backend/config"
	"math"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type ExpenseReq struct {
	GroupID     int     `json:"group_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Splits      []Split `json:"splits"`
}

type Split struct {
	UserID int     `json:"user_id"`
	Amount float64 `json:"amount"`
}

var (
    TotalExpenseAmount = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "hisabkitab_total_expense_amount",
        Help: "The total amount of all expenses created in the app",
    })
)

func init() {
    prometheus.MustRegister(TotalExpenseAmount)
}


func CreateExpense(c *fiber.Ctx) error {
	uid := c.Locals("user_id")
	if uid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	payerID := uid.(int)

	req := new(ExpenseReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "tx error"})
	}

	TotalExpenseAmount.Add(req.Amount)

	defer tx.Rollback()

	// Insert expense
	var expenseID int
	err = tx.QueryRow(
		`INSERT INTO expenses (group_id,payer_id,amount,description)
		 VALUES ($1,$2,$3,$4)
		 RETURNING id`,
		req.GroupID,
		payerID,
		req.Amount,
		req.Description,
	).Scan(&expenseID)

	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error":"expense insert failed"})
	}

	// Decide splits

	var splits []Split

	if len(req.Splits) > 0 {

		// CUSTOM SPLIT
		total := 0.0
		for _, s := range req.Splits {
			total += s.Amount
		}

		if math.Abs(total-req.Amount) > 0.01 {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{"error":"split total mismatch"})
		}

		splits = req.Splits

	} else {

		// EQUAL SPLIT
		rows, _ := tx.Query(
			`SELECT user_id FROM group_members WHERE group_id=$1`,
			req.GroupID,
		)

		defer rows.Close()

		var members []int
		for rows.Next() {
			var id int
			rows.Scan(&id)
			members = append(members,id)
		}

		share := req.Amount / float64(len(members))

		for _, m := range members {
			splits = append(splits, Split{
				UserID: m,
				Amount: share,
			})
		}
	}

	// Insert splits

	for _, s := range splits {
		_, err := tx.Exec(
			`INSERT INTO splits (expense_id,user_id,amount_owed)
			 VALUES ($1,$2,$3)`,
			expenseID,
			s.UserID,
			s.Amount,
		)

		if err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error":"split insert failed"})
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "commit failed"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":    "Expense created",
		"expense_id": expenseID,
	})
}
