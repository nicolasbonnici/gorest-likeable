package likeable

import (
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/filter"
	"github.com/nicolasbonnici/gorest/pagination"
	"github.com/nicolasbonnici/gorest/response"
)

type LikeResource struct {
	DB                 database.Database
	CRUD               *crud.CRUD[Like]
	Config             *Config
	PaginationLimit    int
	PaginationMaxLimit int
}

func RegisterLikeRoutes(app *fiber.App, db database.Database, config *Config) {
	res := &LikeResource{
		DB:                 db,
		CRUD:               crud.New[Like](db),
		Config:             config,
		PaginationLimit:    config.PaginationLimit,
		PaginationMaxLimit: config.MaxPaginationLimit,
	}

	app.Get("/likes", res.List)
	app.Get("/likes/:id", res.Get)
	app.Post("/likes", res.Create)
	app.Put("/likes/:id", res.Update)
	app.Delete("/likes/:id", res.Delete)
}

func (r *LikeResource) List(c *fiber.Ctx) error {
	limit := pagination.ParseIntQuery(c, "limit", r.PaginationLimit, r.PaginationMaxLimit)
	page := pagination.ParseIntQuery(c, "page", 1, 10000)
	page = max(page, 1)
	offset := (page - 1) * limit
	includeCount := c.Query("count", "true") != "false"

	// Field mapping: JSON field name -> DB column name
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

	queryParams := make(url.Values)
	for key, value := range c.Context().QueryArgs().All() {
		queryParams.Add(string(key), string(value))
	}

	filters := filter.NewFilterSetWithMapping(fieldMapping, r.DB.Dialect())
	if err := filters.ParseFromQuery(queryParams); err != nil {
		return pagination.SendPaginatedError(c, 400, err.Error())
	}

	ordering := filter.NewOrderSetWithMapping(fieldMapping)
	if err := ordering.ParseFromQuery(queryParams); err != nil {
		return pagination.SendPaginatedError(c, 400, err.Error())
	}

	filterOrderClauses := ordering.OrderClauses()
	orderByClauses := make([]crud.OrderByClause, len(filterOrderClauses))
	for i, o := range filterOrderClauses {
		orderByClauses[i] = crud.OrderByClause{
			Column:    o.Column,
			Direction: o.Direction,
		}
	}

	result, err := r.CRUD.GetAllPaginated(c.Context(), crud.PaginationOptions{
		Limit:        limit,
		Offset:       offset,
		IncludeCount: includeCount,
		Conditions:   filters.Conditions(),
		OrderBy:      orderByClauses,
	})
	if err != nil {
		return pagination.SendPaginatedError(c, 500, err.Error())
	}

	return pagination.SendHydraCollection(c, result.Items, result.Total, limit, page, r.PaginationLimit)
}

func (r *LikeResource) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	item, err := r.CRUD.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	return response.SendFormatted(c, 200, item)
}

func (r *LikeResource) Create(c *fiber.Ctx) error {
	var req CreateLikeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate request
	if err := req.Validate(r.Config); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	var item Like
	item.Id = uuid.New().String() // Generate UUID before insert
	item.LikeableId = req.LikeableId
	item.Likeable = req.Likeable
	item.LikedId = req.LikedId
	item.LikedAt = time.Now()

	// Try to get user ID from context (set by auth middleware if available)
	if userID := extractUserIDFromContext(c); userID != "" {
		item.LikerId = &userID
	} else {
		// Store IP and User-Agent for anonymous likes
		ip := c.IP()
		ua := c.Get("User-Agent")
		item.IpAddress = &ip
		item.UserAgent = &ua
	}

	ctx := c.Context()
	if err := r.CRUD.Create(ctx, item); err != nil {
		// Check if it's a duplicate like error
		errMsg := err.Error()
		if strings.Contains(errMsg, "UNIQUE constraint") ||
		   strings.Contains(errMsg, "duplicate key") ||
		   strings.Contains(errMsg, "violates unique constraint") {
			return c.Status(409).JSON(fiber.Map{"error": "Already liked"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	created, err := r.CRUD.GetByID(ctx, item.Id)
	if err != nil {
		return response.SendFormatted(c, 201, item)
	}

	return response.SendFormatted(c, 201, created)
}

func (r *LikeResource) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get existing like
	existing, err := r.CRUD.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	// Update liked_at timestamp
	now := time.Now()
	existing.LikedAt = now
	existing.UpdatedAt = &now

	if err := r.CRUD.Update(c.Context(), id, *existing); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return response.SendFormatted(c, 200, existing)
}

func (r *LikeResource) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := r.CRUD.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}

// extractUserIDFromContext tries to extract user ID from context
// This allows optional integration with auth middleware that sets user ID in context
// Example: c.Locals("user_id")
func extractUserIDFromContext(c *fiber.Ctx) string {
	// Try common context keys used by auth middleware
	if userID := c.Locals("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	if userID := c.Locals("userId"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}
