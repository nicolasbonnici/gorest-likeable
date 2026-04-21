package likeable

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/database"
)

func RegisterRoutes(router fiber.Router, db database.Database, config *Config) {
	RegisterLikeRoutes(router, db, config)
}
