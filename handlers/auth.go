package handlers

import (
	"context"
	"go_backend/config"
    "github.com/gofiber/fiber/v2"
    "golang.org/x/crypto/bcrypt"
    "go_backend/models" // Adjust based on your module name
)

func Signup(c *fiber.Ctx) error {
    req := new(models.SignupRequest)

    // Step 1: Parse the body
    if err := c.BodyParser(req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Review your input"})
    }

	if req.Password != req.ConfirmPassword {
		return c.Status(400).JSON(fiber.Map{"error": "Password and Confirm Password do not match"})
	}

    // Convert string to bytes, then hash it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not hash password"})
	}

	newUser := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	_, err = config.UserCollection.InsertOne(context.Background(), newUser)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create user"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "User registered successfully!"})
}