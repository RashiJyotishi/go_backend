package handlers

import (
	"go_backend/config"
	"go_backend/models"
	"os"
	"time"
	// "log"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func Signup(c *fiber.Ctx) error {
	req := new(models.SignupRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	// log.Println("Signup request:", req.ConfirmPassword)

	if req.Password != req.ConfirmPassword {
		return c.Status(400).JSON(fiber.Map{"error": "Passwords do not match"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not hash password"})
	}

	_, err = config.DB.Exec(
		`INSERT INTO users (username,email,password)
		 VALUES ($1,$2,$3)`,
		req.Username,
		req.Email,
		string(hashedPassword),
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "User already exists"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "User registered successfully"})
}

func Login(c *fiber.Ctx) error {
	req := new(models.LoginRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	var user models.User

	err := config.DB.QueryRow(
		"SELECT id,password FROM users WHERE email=$1",
		req.Email,
	).Scan(&user.ID, &user.Password)

	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	goToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{"token": goToken})
}

func GetDashboard(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Welcome to the dashboard!"})
}
