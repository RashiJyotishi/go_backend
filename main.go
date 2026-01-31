package main

import (
	"github.com/joho/godotenv"
	"go_backend/config"
	"go_backend/handlers"
	"go_backend/middleware"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	_ = godotenv.Load()

	config.ConnectDB()

	app := fiber.New()

	// ✅ PUBLIC ROUTES
	app.Post("/api/signup", handlers.Signup)
	app.Post("/api/login", handlers.Login)

	// 🔐 PROTECTED ROUTES
	protected := app.Group("/api", middleware.AuthRequired)
	protected.Get("/dashboard", handlers.GetDashboard)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server is running")
	})

	log.Fatal(app.Listen(":8080"))
}
