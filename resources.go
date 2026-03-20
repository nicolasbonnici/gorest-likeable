package likeable

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/processor"
)

type LikeResource struct {
	processor processor.Processor[Like, LikeCreateDTO, LikeUpdateDTO, LikeResponseDTO]
}

func RegisterLikeRoutes(app *fiber.App, db database.Database, config *Config) {
	likeCRUD := crud.New[Like](db)
	hooks := NewLikeHooks(db, config)
	converter := &LikeConverter{}
	errorHandler := &LikeErrorHandler{}

	fieldMapping := map[string]string{
		"id":         "id",
		"likerId":    "liker_id",
		"likedId":    "liked_id",
		"likeableId": "likeable_id",
		"likeable":   "likeable",
		"ipAddress":  "ip_address",
		"userAgent":  "user_agent",
		"likedAt":    "liked_at",
		"updatedAt":  "updated_at",
		"createdAt":  "created_at",
	}

	proc := processor.New(processor.ProcessorConfig[Like, LikeCreateDTO, LikeUpdateDTO, LikeResponseDTO]{
		DB:                 db,
		CRUD:               likeCRUD,
		Converter:          converter,
		PaginationLimit:    config.PaginationLimit,
		PaginationMaxLimit: config.MaxPaginationLimit,
		FieldMap:           fieldMapping,
		AllowedFields:      []string{"id", "likerId", "likedId", "likeableId", "likeable", "ipAddress", "userAgent", "likedAt", "updatedAt", "createdAt"},
		ErrorHandler:       errorHandler,
	}).
		WithCreateHook(hooks.CreateHook).
		WithUpdateHook(hooks.UpdateHook).
		WithDeleteHook(hooks.DeleteHook).
		WithGetAllHook(hooks.GetAllHook)

	res := &LikeResource{
		processor: proc,
	}

	app.Get("/likes", res.GetAll)
	app.Get("/likes/:id", res.GetByID)
	app.Post("/likes", res.Create)
	app.Put("/likes/:id", res.Update)
	app.Delete("/likes/:id", res.Delete)
}

func (r *LikeResource) Create(c *fiber.Ctx) error {
	return r.processor.Create(c)
}

func (r *LikeResource) GetByID(c *fiber.Ctx) error {
	return r.processor.GetByID(c)
}

func (r *LikeResource) GetAll(c *fiber.Ctx) error {
	return r.processor.GetAll(c)
}

func (r *LikeResource) Update(c *fiber.Ctx) error {
	return r.processor.Update(c)
}

func (r *LikeResource) Delete(c *fiber.Ctx) error {
	return r.processor.Delete(c)
}
