package handlers

import (
	"os"
	"context"
	"go_backend/config"
	"go_backend/models" // Adjust based on your module name
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

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

func Login(c *fiber.Ctx) error {
	log.Println("LOGIN HANDLER HIT")
	req := new(models.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "review your input"})
	}
	var user models.User
	filter := bson.M{"email": req.Email}
	log.Println("Attempting login for email:", req.Email)
	err := config.UserCollection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
    }

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
    }

	claims := jwt.MapClaims{
        "user_id": user.ID.Hex(),
        "exp":     time.Now().Add(time.Hour * 72).Unix(), // Expires in 72 hours
    }
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	log.Println("Generating token for user ID:", user.ID.Hex())

	go_token, err := token.SignedString(jwtSecret)
    if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "could not generate token",
		})
    }
	log.Println("Login successful:", user.ID.Hex())

	return c.JSON(fiber.Map{"token": go_token})

}

func GetDashboard(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"message": "Welcome to the dashboard!"})
}