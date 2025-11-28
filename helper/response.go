package helper

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

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

func GetIntQuery(c *fiber.Ctx, key string, defaultValue int) int {
	val := c.Query(key)

	if val == "" {
		return defaultValue
	}

	num, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}

	return num
}
func GetUserID(c *fiber.Ctx) string {
    // Unit test pakai Header
    if h := c.Get("user_id"); h != "" {
        return h
    }

    // Real API pakai Locals
    if v := c.Locals("user_id"); v != nil {
        return v.(string)
    }

    return ""
}
