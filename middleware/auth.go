package middleware

import (
	"os"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"strings"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func AuthRequired(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	// log.Print("Auth Header:")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing or invalid token",
		})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	// log.Print("Token String: ")
	go_token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !go_token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}
// log.Print("Token Validated: ")

	claims := go_token.Claims.(jwt.MapClaims)

	// THIS IS IMPORTANT
	userID := int(claims["user_id"].(float64))

	// attach user_id to request
	c.Locals("user_id", userID)
// log.Print("User ID from token: ")
	return c.Next()
}
