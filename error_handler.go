package likeable

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

type LikeErrorHandler struct{}

func (h *LikeErrorHandler) HandleError(c *fiber.Ctx, err error, operation string) error {
	if operation == "create" {
		errMsg := err.Error()
		if strings.Contains(errMsg, "UNIQUE constraint") ||
			strings.Contains(errMsg, "duplicate key") ||
			strings.Contains(errMsg, "violates unique constraint") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Already liked"})
		}
	}

	if fiberErr, ok := err.(*fiber.Error); ok {
		return c.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
	}

	switch operation {
	case "parse":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	case "validate":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case "getById", "getAll":
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows in result set") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	case "update", "delete":
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows in result set") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}
