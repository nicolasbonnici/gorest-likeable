package likeable

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/response"
)

type LikeErrorHandler struct{}

func (h *LikeErrorHandler) HandleError(c fiber.Ctx, err error, operation string) error {
	if operation == "create" && crud.IsDuplicateError(err) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Already liked"})
	}

	if fiberErr, ok := err.(*fiber.Error); ok {
		msg := fiberErr.Message
		if fiberErr.Code >= 500 {
			msg = "Internal server error"
		}
		return c.Status(fiberErr.Code).JSON(fiber.Map{"error": msg})
	}

	// Checked ahead of the switch because it is the same answer everywhere,
	// and because crud.ErrInvalidID wraps sql.ErrNoRows: a like id that is not
	// a UUID has not found a row, and saying so is what stops the driver's
	// `invalid input syntax for type uuid: "0" (SQLSTATE 22P02)` from being
	// repeated to the caller.
	if crud.IsNotFoundError(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not found"})
	}

	// Like.Id is declared `string` over a uuid column, so crud.checkID cannot
	// reject a malformed id locally and the driver is what refuses it. Answer
	// it the same way a missing row is answered, or the two are
	// distinguishable from outside and the id column's type is disclosed.
	if crud.IsInvalidIDError(err) {
		switch operation {
		case "getById", "update", "delete":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not found"})
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid identifier in request"})
		}
	}

	switch operation {
	case "parse":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	case "validate":
		// The only branch that repeats a message, and only one a validator
		// wrote: SafeMessage substitutes the fallback for anything carrying
		// driver or runtime text.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": response.SafeMessage(err, "Validation failed")})
	case "getById", "getAll", "update", "delete":
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
}
