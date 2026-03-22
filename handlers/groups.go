package handlers

import (
	"math/rand"
	"strconv"
	"time"
	"log"
)
import (
	"go_backend/config"

	"github.com/gofiber/fiber/v2"
)

func GetUserGroups(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	rows, err := config.DB.Query(`
		SELECT g.id, g.name, g.join_code
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1
	`, userID)


	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error"})
	}

	defer rows.Close()

	var groups []fiber.Map

	for rows.Next() {
		var id int
		var name string
		var code string

		rows.Scan(&id, &name, &code)

		groups = append(groups, fiber.Map{
			"id": id,
			"name": name,
			"join_code": code,
		})
	}

	return c.JSON(groups)
}
func generateJoinCode() string {
	rand.Seed(time.Now().UnixNano())
	return strconv.Itoa(100000 + rand.Intn(900000)) // 6 digit
}
func CreateGroup(c *fiber.Ctx) error {
	log.Println("CREATE GROUP HIT")

	uid := c.Locals("user_id")
	log.Println("USER ID:", uid)
	// uid := c.Locals("user_id")
	log.Println("LOCALS:", c.Locals("user_id"))
	if uid == nil {
		return c.Status(401).JSON(fiber.Map{"error":"unauthorized"})
	}
	userID := uid.(int)


	type Req struct {
		Name string `json:"name"`
	}

	req := new(Req)
	c.BodyParser(req)

	code := generateJoinCode()

	var groupID int

	err := config.DB.QueryRow(
		`INSERT INTO groups (name, join_code)
		 VALUES ($1,$2)
		 RETURNING id`,
		req.Name,
		code,
	).Scan(&groupID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error":"could not create group"})
	}

	// creator joins automatically
	config.DB.Exec(
		`INSERT INTO group_members (group_id,user_id)
		 VALUES ($1,$2)`,
		groupID,
		userID,
	)

	return c.JSON(fiber.Map{
		"group_id": groupID,
		"join_code": code,
	})
}

func JoinGroup(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	type Req struct {
		Code string `json:"code"`
	}

	req := new(Req)
	c.BodyParser(req)

	var groupID int

	err := config.DB.QueryRow(
		`SELECT id FROM groups WHERE join_code=$1`,
		req.Code,
	).Scan(&groupID)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error":"invalid code"})
	}

	_, err = config.DB.Exec(
		`INSERT INTO group_members (group_id,user_id)
		 VALUES ($1,$2)
		 ON CONFLICT DO NOTHING`,
		groupID,
		userID,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error":"could not join"})
	}

	return c.JSON(fiber.Map{"message":"joined group"})
}

func GetGroupMembers(c *fiber.Ctx) error {
	log.Println("GET GROUP MEMBERS HIT")
    groupID := c.Params("id")
	log.Println("GET GROUP MEMBERS HIT, GROUP ID:", groupID)
    rows, err := config.DB.Query("SELECT u.id, u.username FROM group_members gm JOIN users u ON gm.user_id = u.id WHERE gm.group_id = $1", groupID)
    if err != nil { return c.Status(500).JSON(fiber.Map{"error": "db error"}) }
    defer rows.Close()

	type Member struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	}
    var members []Member
	log.Println("GROUP ID:", groupID)
    for rows.Next() {
        var m Member
        if err := rows.Scan(&m.ID, &m.Username); err == nil {
            members = append(members, m)
        }
    }
    if members == nil { members = []Member{} }
    return c.JSON(members)
}

