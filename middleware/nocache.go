package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
)

var (
	epoch = time.Unix(0, 0).Format(time.RFC1123)

	noCacheHeaders = map[string]string{
		"Expires":         epoch,
		"Cache-Control":   "no-cache, private, max-age=0",
		"Pragma":          "no-cache",
		"X-Accel-Expires": "0",
	}
)

func NoCache() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := c.Response()
			for key, val := range noCacheHeaders {
				res.Header().Set(key, val)
			}
			return next(c)
		}
	}
}
