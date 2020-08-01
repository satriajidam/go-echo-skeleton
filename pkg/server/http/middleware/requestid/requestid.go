package requestid

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// New initializes the request ID middleware.
func New() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// Get id from request
			rid := ctx.Request().Header.Get(echo.HeaderXRequestID)

			if rid == "" {
				rid = uuid.New().String()
			}

			ctx.Set(echo.HeaderXRequestID, rid)
			ctx.Request().Header.Set(echo.HeaderXRequestID, rid)
			ctx.Response().Header().Set(echo.HeaderXRequestID, rid)

			return next(ctx)
		}
	}
}
