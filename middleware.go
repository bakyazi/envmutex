package envmutex

import "github.com/labstack/echo/v4"

func WithUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := c.Request().Cookie("env-user")
		if err != nil {
			return Login(c)
		}
		return next(c)
	}
}
