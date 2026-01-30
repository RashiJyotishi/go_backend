package main

import (
    "go_backend/config"
    "go_backend/handlers"
    "github.com/gofiber/fiber/v2"
    "log"
)

func main() {

    config.ConnectDB()
    app := fiber.New()
    app.Post("/signup", handlers.Signup)

    // A simple health check route
    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Server is up and running!")
    })

    // Start the server on port 8080
    log.Fatal(app.Listen(":8080"))
}