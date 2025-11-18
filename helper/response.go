package helper

import "github.com/gofiber/fiber/v2"

// Standard API response format
func APIResponse(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"status":  status,
		"message": message,
		"data":    data,
	})
}

// Common reusable responses
func BadRequest(c *fiber.Ctx, msg string) error {
	return APIResponse(c, fiber.StatusBadRequest, msg, nil)
}

func Unauthorized(c *fiber.Ctx, msg string) error {
	return APIResponse(c, fiber.StatusUnauthorized, msg, nil)
}

func Forbidden(c *fiber.Ctx, msg string) error {
	return APIResponse(c, fiber.StatusForbidden, msg, nil)
}

func NotFound(c *fiber.Ctx, msg string) error {
	return APIResponse(c, fiber.StatusNotFound, msg, nil)
}

func Conflict(c *fiber.Ctx, msg string) error {
	return APIResponse(c, fiber.StatusConflict, msg, nil)
}

func Unprocessable(c *fiber.Ctx, msg string) error {
	return APIResponse(c, fiber.StatusUnprocessableEntity, msg, nil)
}

func InternalError(c *fiber.Ctx, msg string) error {
	return APIResponse(c, fiber.StatusInternalServerError, msg, nil)
}
