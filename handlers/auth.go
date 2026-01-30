package handlers

import (
	"context"
	"time"
    "github.com/golang-jwt/jwt/v5"
	"go_backend/config"
    "github.com/gofiber/fiber/v2"
    "golang.org/x/crypto/bcrypt"
	"go_backend/models" // Adjust based on your module name
	"go.mongodb.org/mongo-driver/bson"
)

var jwtSecret = []byte("JWT_secret_key")

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
	req := new(models.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "review your input"})
	}
	var user models.User
	filter := bson.M{"email": req.Email}
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

	go_token, err := token.SignedString(jwtSecret)
    if err != nil {
        return c.SendStatus(fiber.StatusInternalServerError)
    }
	return c.JSON(fiber.Map{"token": go_token})

	// return c.Status(200).JSON(fiber.Map{"message": "Login successful!"})
}