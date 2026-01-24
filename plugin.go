package likeable

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest-likeable/migrations"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/plugin"
)

type LikeablePlugin struct {
	config Config
	db     database.Database
}

func NewPlugin() plugin.Plugin {
	return &LikeablePlugin{}
}

func (p *LikeablePlugin) Name() string {
	return "likeable"
}

func (p *LikeablePlugin) Dependencies() []string {
	return []string{"auth"}
}

func (p *LikeablePlugin) Initialize(config map[string]interface{}) error {
	p.config = DefaultConfig()

	if db, ok := config["database"].(database.Database); ok {
		p.db = db
		p.config.Database = db
	}

	if allowedTypes, ok := config["allowed_types"].([]interface{}); ok {
		types := make([]string, 0, len(allowedTypes))
		for _, t := range allowedTypes {
			if str, ok := t.(string); ok {
				types = append(types, str)
			}
		}
		if len(types) > 0 {
			p.config.AllowedTypes = types
		}
	}

	if paginationLimit, ok := config["pagination_limit"].(int); ok {
		p.config.PaginationLimit = paginationLimit
	}

	if maxPaginationLimit, ok := config["max_pagination_limit"].(int); ok {
		p.config.MaxPaginationLimit = maxPaginationLimit
	}

	if enableUserLikes, ok := config["enable_user_likes"].(bool); ok {
		p.config.EnableUserLikes = enableUserLikes
	}

	return p.config.Validate()
}

func (p *LikeablePlugin) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

func (p *LikeablePlugin) SetupEndpoints(app *fiber.App) error {
	if p.db == nil {
		return nil
	}

	RegisterRoutes(app, p.db, &p.config)
	return nil
}

func (p *LikeablePlugin) MigrationSource() interface{} {
	return migrations.GetMigrations()
}

func (p *LikeablePlugin) MigrationDependencies() []string {
	return []string{"auth"}
}

func (p *LikeablePlugin) GetOpenAPIResources() []plugin.OpenAPIResource {
	return []plugin.OpenAPIResource{{
		Name:          "like",
		PluralName:    "likes",
		BasePath:      "/likes",
		Tags:          []string{"Likes"},
		ResponseModel: Like{},
		CreateModel:   CreateLikeRequest{},
		UpdateModel:   UpdateLikeRequest{},
		Description:   "Like/unlike system for posts and comments",
	}}
}
