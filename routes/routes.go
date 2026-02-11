package routes

import (
	"go_backend/handlers"
	"go_backend/middleware"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/chat/:id", websocket.New(handlers.WebsocketHandler))

	api := app.Group("/api")

	api.Post("/signup", handlers.Signup)
	api.Post("/login", handlers.Login)

	protected := api.Group("/", middleware.AuthRequired)

	protected.Get("/dashboard", handlers.GetDashboard)

	protected.Get("/groups", handlers.GetUserGroups)
	protected.Post("/create-group", handlers.CreateGroup)
	protected.Post("/join-group", handlers.JoinGroup)

	protected.Post("/expenses", handlers.CreateExpense)
	protected.Post("/settlements", handlers.CreateSettlement)

	protected.Get("/groups/:id/simplify", handlers.SimplifyGroup)
	protected.Get("/groups/:id/activity", handlers.GetGroupActivity)
	protected.Get("/groups/:id/expenses", handlers.GetGroupExpenses)
}