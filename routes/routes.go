package routes

import (
	"go_backend/handlers"
	"github.com/gofiber/fiber/v2"
)

func Route(app *fiber.App) {
    api := app.Group("/api")

    // Public Routes
    api.Post("/signup", handlers.Signup)
    api.Post("/login", handlers.Login)

    // Protected Routes (User must be logged in)
    // We apply the middleware here
    // tasks := api.Group("/tasks", AuthRequired)
    // tasks.Post("/submit", handlers.SubmitTask)
}