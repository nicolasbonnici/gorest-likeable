package likeable

import (
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	auth "github.com/nicolasbonnici/gorest-auth"
	"github.com/nicolasbonnici/gorest-rbac"
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
	Voter              rbac.Voter
}

func RegisterLikeRoutes(app *fiber.App, db database.Database, config *Config) {
	rbacConfig := rbac.Config{
		DefaultPolicy: rbac.DenyAll,
		SuperuserRole: "admin",
		RoleHierarchy: map[string][]string{
			"writer":    {"moderator"},
			"moderator": {"reader"},
		},
		CacheEnabled: true,
		StrictMode:   false,
	}

	voter, err := rbac.NewVoter(rbacConfig)
	if err != nil {
		panic("failed to create RBAC voter: " + err.Error())
	}

	roleProvider := rbac.NewFiberRoleProvider("user_roles", "user_id")

	res := &LikeResource{
		DB:                 db,
		CRUD:               crud.New[Like](db),
		Config:             config,
		PaginationLimit:    config.PaginationLimit,
		PaginationMaxLimit: config.MaxPaginationLimit,
		Voter:              voter,
	}

	app.Get("/likes", res.List)
	app.Get("/likes/:id", res.Get)
	app.Post("/likes", rbac.RequireRole(voter, roleProvider, "reader"), res.Create)
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

	result, err := r.CRUD.GetAllPaginated(auth.Context(c), crud.PaginationOptions{
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
	item, err := r.CRUD.GetByID(auth.Context(c), id)
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

	// Require authentication
	user := auth.GetAuthenticatedUser(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Authentication required"})
	}

	ctx := auth.Context(c)

	var item Like
	item.Id = uuid.New().String() // Generate UUID before insert
	item.LikerId = &user.UserID
	item.LikeableId = req.LikeableId
	item.Likeable = req.Likeable
	item.LikedId = req.LikedId
	item.LikedAt = time.Now()

	// Validate RBAC permissions
	if err := r.Voter.ValidateWrite(ctx, &item); err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

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
	ctx := auth.Context(c)

	// Get existing like
	existing, err := r.CRUD.GetByID(ctx, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	// Update liked_at timestamp
	now := time.Now()
	existing.LikedAt = now
	existing.UpdatedAt = &now

	if err := r.CRUD.Update(ctx, id, *existing); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return response.SendFormatted(c, 200, existing)
}

func (r *LikeResource) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := auth.Context(c)

	// Get existing like to check ownership
	existing, err := r.CRUD.GetByID(ctx, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	// Check ownership
	user := auth.GetAuthenticatedUser(c)
	if user == nil || existing.LikerId == nil || *existing.LikerId != user.UserID {
		return c.Status(403).JSON(fiber.Map{"error": "You can only delete your own likes"})
	}

	if err := r.CRUD.Delete(ctx, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}

