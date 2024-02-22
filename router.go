package envmutex

import (
	"github.com/bakyazi/envmutex/middleware"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bakyazi/envmutex/components"
	"github.com/bakyazi/envmutex/service"
	"github.com/bakyazi/envmutex/service/sheets"
	"github.com/labstack/echo/v4"
)

type Router struct {
	s service.Service
}

func Init(e *echo.Echo) {
	googleSheetsID := os.Getenv("GOOGLE_SHEET_ID")

	sheetService, err := sheets.NewService(googleSheetsID)
	if err != nil {
		log.Fatal(err)
	}

	e.Use(middleware.NoCache())
	router := NewRouter(sheetService)
	e.POST("/login", router.Login)
	e.POST("/logout", router.Logout)

	apiGroup := e.Group("")
	apiGroup.Use(WithUser)
	apiGroup.GET("/", router.Home)
	apiGroup.GET("/:env/lock", router.LockEnv)
	apiGroup.GET("/:env/release", router.ReleaseEnv)
}

func NewRouter(service service.Service) *Router {
	return &Router{s: service}
}

func (r *Router) Home(c echo.Context) error {
	user, err := c.Request().Cookie("env-user")
	if err != nil || user == nil {
		return err
	}
	envs, err := r.s.GetEnvironments()
	if err != nil {
		return err
	}
	return components.Home(user.Value, envs).Render(c.Request().Context(), c.Response().Writer)
}

func (r *Router) LockEnv(c echo.Context) error {
	user, err := c.Request().Cookie("env-user")
	if err != nil || user == nil {
		return err
	}

	env := c.Param("env")
	err = r.s.LockEnvironment(env, user.Value)
	if err != nil {
		return err
	}
	return r.Home(c)
}

func (r *Router) ReleaseEnv(c echo.Context) error {
	user, err := c.Request().Cookie("env-user")
	if err != nil || user == nil {
		return err
	}

	env := c.Param("env")
	err = r.s.ReleaseEnvironment(env, user.Value)
	if err != nil {
		return err
	}
	return r.Home(c)
}

func (r *Router) Login(c echo.Context) error {
	type loginParams struct {
		Name string `json:"name" form:"name"`
	}
	params := new(loginParams)
	if err := c.Bind(params); err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:  "env-user",
		Value: params.Name,
	})

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func (r *Router) Logout(c echo.Context) error {
	cookie, err := c.Cookie("env-user")
	if err != nil || cookie == nil {
		return err
	}
	cookie.Name = "env-user"
	cookie.Expires = time.Now()
	c.SetCookie(cookie)
	c.SetCookie(&http.Cookie{
		Name:  "Cache-Control",
		Value: "no-cache",
	})
	return c.Redirect(http.StatusMovedPermanently, "/")
}

func Login(c echo.Context) error {
	return components.Login().Render(c.Request().Context(), c.Response().Writer)
}
