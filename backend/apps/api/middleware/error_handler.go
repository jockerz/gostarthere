package middleware

import "github.com/gofiber/fiber/v3"

func ErrorHandler(ctx fiber.Ctx, err error) error {
	// code := fiber.StatusInternalServerError

	// var e

	// err = ctx.Status(code)

	return nil
}
