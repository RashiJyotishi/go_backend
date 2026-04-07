package main

import (
    "github.com/joho/godotenv"
    "go_backend/config"
    "go_backend/handlers"
    "go_backend/middleware"
    "github.com/gofiber/fiber/v2"
    "log"
    "github.com/gofiber/contrib/websocket"
    "github.com/gofiber/adaptor/v2"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
    _ = godotenv.Load()

    config.ConnectDB()

    app := fiber.New()
    // This creates the /metrics endpoint for Prometheus to "scrape"
    app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

    app.Use("/ws", func(c *fiber.Ctx) error {
        if websocket.IsWebSocketUpgrade(c) {
            c.Locals("allowed", true)
            return c.Next()
        }
        return fiber.ErrUpgradeRequired
    })
    app.Use(cors.New(cors.Config{
        AllowOrigins:     "http://localhost:8001,http://localhost:8000", // yaad rakhna -- update this in production
        AllowMethods:     "GET,POST,OPTIONS",
        AllowHeaders:     "Accept,Authorization,Content-Type",
        AllowCredentials: true,
    }))

    app.Get("/ws/chat/:id", websocket.New(handlers.WebsocketHandler))
    // Public Routes
    app.Post("/api/signup", handlers.Signup)
    app.Post("/api/login", handlers.Login)

    // Protected Routes
    protected := app.Group("/api", middleware.AuthRequired)

    protected.Get("/dashboard", handlers.GetDashboard)
    protected.Get("/groups", handlers.GetUserGroups)
    protected.Post("/create-group", handlers.CreateGroup)
    protected.Post("/join-group", handlers.JoinGroup)

    // Expenses & Settlements
    protected.Post("/expenses", handlers.CreateExpense)
    protected.Post("/settlements", handlers.CreateSettlement)

    // Group Data
    protected.Get("/groups/:id/simplify", handlers.SimplifyGroup)
    protected.Get("/groups/:id/activity", handlers.GetGroupActivity)
    protected.Get("/groups/:id/expenses", handlers.GetGroupExpenses)
    protected.Get("/groups/:id/members", handlers.GetGroupMembers)
    protected.Get("/groups/:id/chat-pagination", handlers.GetChatPagination)
    // protected.Get("/groups/:id/chat-history", handlers.GetChatHistory)

    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Server is running")
    })

    log.Fatal(app.Listen(":8080"))
}