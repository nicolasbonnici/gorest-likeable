package likeable

import (
	"github.com/gofiber/fiber/v3"
	auth "github.com/nicolasbonnici/gorest/auth"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/processor"
)

type LikeResource struct {
	processor processor.Processor[Like, LikeCreateDTO, LikeUpdateDTO, LikeResponseDTO]
	service   *LikeService
}

func RegisterLikeRoutes(router fiber.Router, db database.Database, config *Config) {
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
		service:   NewLikeService(db),
	}

	router.Get("/likes", res.GetAll)
	// Static routes are registered before the "/likes/:id" parameter route so
	// they are not shadowed by it.
	router.Get("/likes/count", res.Count)
	router.Post("/likes/state", res.State)
	router.Get("/likes/:id", res.GetByID)
	router.Post("/likes", res.Create)
	router.Put("/likes/:id", res.Update)
	router.Delete("/likes/:id", res.Delete)
}

func (r *LikeResource) Create(c fiber.Ctx) error {
	return r.processor.Create(c)
}

func (r *LikeResource) GetByID(c fiber.Ctx) error {
	return r.processor.GetByID(c)
}

func (r *LikeResource) GetAll(c fiber.Ctx) error {
	return r.processor.GetAll(c)
}

func (r *LikeResource) Update(c fiber.Ctx) error {
	return r.processor.Update(c)
}

func (r *LikeResource) Delete(c fiber.Ctx) error {
	return r.processor.Delete(c)
}

func (r *LikeResource) Count(c fiber.Ctx) error {
	likeableType := c.Query("likeable")
	likeableID := c.Query("likeableId")
	if likeableType == "" || likeableID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "likeable and likeableId are required")
	}

	ctx := auth.Context(c)
	count, err := r.service.Count(ctx, likeableType, likeableID)
	if err != nil {
		return err
	}

	liked := false
	if user := auth.GetAuthenticatedUser(c); user != nil {
		states, err := r.service.LikedByBatch(ctx, user.UserID, likeableType, []string{likeableID})
		if err != nil {
			return err
		}
		liked = states[likeableID]
	}

	return c.JSON(LikeCountResponseDTO{
		Likeable:   likeableType,
		LikeableId: likeableID,
		Count:      count,
		Liked:      liked,
	})
}

func (r *LikeResource) State(c fiber.Ctx) error {
	var req LikeStateRequestDTO
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if req.Likeable == "" {
		return fiber.NewError(fiber.StatusBadRequest, "likeable is required")
	}

	ctx := auth.Context(c)
	counts, err := r.service.CountBatch(ctx, req.Likeable, req.LikeableIds)
	if err != nil {
		return err
	}

	likerID := ""
	if user := auth.GetAuthenticatedUser(c); user != nil {
		likerID = user.UserID
	}
	liked, err := r.service.LikedByBatch(ctx, likerID, req.Likeable, req.LikeableIds)
	if err != nil {
		return err
	}

	states := make(map[string]LikeStateDTO, len(req.LikeableIds))
	for _, id := range req.LikeableIds {
		states[id] = LikeStateDTO{Count: counts[id], Liked: liked[id]}
	}

	return c.JSON(LikeStateResponseDTO{States: states})
}
