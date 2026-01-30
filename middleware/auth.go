package middleware

import (
	"os"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"strings"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

func AuthRequired(c *fiber.Ctx) error {
	// 1. Get the Authorization header (Expected format: "Bearer <token>")
	authHeader := c.Get("Authorization")

	// 2. Check if the header is empty or doesn't start with "Bearer "
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing or invalid token",
		})
	}

	// 3. Extract the actual token string
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// 4. Parse and Validate the token
	go_token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !go_token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// 5. Success! Move to the next function (the actual Task Handler)
	return c.Next()
}