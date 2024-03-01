package envmutex

import (
	"errors"
	"github.com/bakyazi/envmutex/service"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"os"
	"time"
)

func WithUser(authService service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := func() error {
				token, err := c.Request().Cookie("token")
				if err != nil {
					return err
				}

				t, err := jwt.Parse(token.Value, func(t *jwt.Token) (interface{}, error) {
					if jwt.GetSigningMethod("HS256") != t.Method {
						return nil, errors.New("invalid algo")
					}
					return []byte(os.Getenv("JWT_PRIV_KEY")), nil
				})

				err = t.Claims.Valid()
				if err != nil {
					token.Expires = time.Now()
					c.SetCookie(token)
					return errors.New("not valid token")
				}

				mc, ok := t.Claims.(jwt.MapClaims)
				if !ok {
					return errors.New("not valid token")
				}

				name, ok := mc["name"]
				if !ok {
					return errors.New("not valid token")
				}
				c.Set("username", name)
				return nil
			}()
			if err != nil {
				return Login(c)
			}

			return next(c)
		}
	}
}
