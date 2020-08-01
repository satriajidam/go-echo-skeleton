package requestid

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const KeyRequestID = "RequestID"
const HeaderXRequestID = "X-Request-ID"

// New initializes the request ID middleware.
func New() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// Get id from request
			rid := ctx.Request().Header.Get(HeaderXRequestID)

			if rid == "" {
				rid = uuid.New().String()
			}

			ctx.Set(KeyRequestID, rid)
			ctx.Request().Header.Set(HeaderXRequestID, rid)
			ctx.Response().Header().Set(HeaderXRequestID, rid)

			return next(ctx)
		}
	}
}
