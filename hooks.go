package likeable

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	auth "github.com/nicolasbonnici/gorest-auth"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/query"
)

type LikeHooks struct {
	db     database.Database
	config *Config
}

func NewLikeHooks(db database.Database, config *Config) *LikeHooks {
	return &LikeHooks{
		db:     db,
		config: config,
	}
}

func (h *LikeHooks) CreateHook(c *fiber.Ctx, dto LikeCreateDTO, model *Like) error {
	if dto.Likeable == "user" {
		if !h.config.EnableUserLikes {
			return fiber.NewError(400, "user likes are not enabled")
		}
		if dto.LikedId == nil || *dto.LikedId == "" {
			return fiber.NewError(400, "likedId is required when liking a user")
		}
	} else {
		if !h.config.IsAllowedType(dto.Likeable) {
			return fiber.NewError(400, "likeable type is not allowed")
		}
	}

	user := auth.GetAuthenticatedUser(c)
	if user != nil {
		model.LikerId = &user.UserID
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	if ipAddress != "" {
		model.IpAddress = &ipAddress
	}
	if userAgent != "" {
		model.UserAgent = &userAgent
	}

	return nil
}

func (h *LikeHooks) UpdateHook(c *fiber.Ctx, dto LikeUpdateDTO, model *Like) error {
	return fiber.NewError(405, "Method not implemented")
}

func (h *LikeHooks) DeleteHook(c *fiber.Ctx, id any) error {
	ctx := auth.Context(c)

	existing, err := h.getLike(ctx, id)
	if err != nil {
		return fiber.NewError(404, "Not found")
	}

	user := auth.GetAuthenticatedUser(c)
	if user == nil || existing.LikerId == nil || *existing.LikerId != user.UserID {
		return fiber.NewError(403, "You can only delete your own likes")
	}

	return nil
}

func (h *LikeHooks) GetAllHook(c *fiber.Ctx, conditions *[]query.Condition, orderBy *[]crud.OrderByClause) error {
	return nil
}

func (h *LikeHooks) getLike(ctx context.Context, id any) (*Like, error) {
	var like Like
	idStr, ok := id.(string)
	if !ok {
		return nil, errors.New("invalid ID type")
	}

	sql := "SELECT * FROM likes WHERE id = " + h.db.Dialect().Placeholder(1)
	err := h.db.QueryRow(ctx, sql, idStr).Scan(
		&like.Id,
		&like.LikerId,
		&like.LikedId,
		&like.LikeableId,
		&like.Likeable,
		&like.IpAddress,
		&like.UserAgent,
		&like.LikedAt,
		&like.UpdatedAt,
		&like.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &like, nil
}
