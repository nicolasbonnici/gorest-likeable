package likeable

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/database"
)

func RegisterRoutes(app *fiber.App, db database.Database, config *Config) {
	RegisterLikeRoutes(app, db, config)
}
