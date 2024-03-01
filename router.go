package envmutex

import (
	"errors"
	"github.com/bakyazi/envmutex/middleware"
	"google.golang.org/api/option"
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
	a service.AuthService
}

func Init(e *echo.Echo) {
	googleSheetsID := os.Getenv("GOOGLE_SHEET_ID")
	var sheetService *sheets.Service
	var err error
	if os.Getenv("USE_CREDENTIAL_FILE") == "TRUE" {
		sheetService, err = sheets.NewService(googleSheetsID, option.WithCredentialsFile("credentials.json"))
	} else {
		sheetService, err = sheets.NewService(googleSheetsID)
	}
	if err != nil {
		log.Fatal(err)
	}

	e.Use(middleware.NoCache())
	router := NewRouter(sheetService, sheetService)
	e.POST("/login", router.Login)

	apiGroup := e.Group("")
	apiGroup.Use(WithUser(sheetService))
	apiGroup.GET("/", router.Home)
	apiGroup.GET("/:env/lock", router.LockEnv)
	apiGroup.GET("/:env/release", router.ReleaseEnv)
	apiGroup.GET("/reset-password", router.ResetPasswordForm)
	apiGroup.POST("/reset-password", router.ResetPassword)
	apiGroup.POST("/logout", router.Logout)

}

func NewRouter(service service.Service, authService service.AuthService) *Router {
	return &Router{s: service, a: authService}
}

func (r *Router) Home(c echo.Context) error {
	user := c.Get("username").(string)
	envs, err := r.s.GetEnvironments()
	if err != nil {
		return ReturnError(c, 400, err)
	}
	return components.Home(user, envs).Render(c.Request().Context(), c.Response().Writer)
}

func (r *Router) LockEnv(c echo.Context) error {
	user := c.Get("username").(string)

	env := c.Param("env")
	err := r.s.LockEnvironment(env, user)
	if err != nil {
		return ReturnError(c, 400, err)
	}
	return r.Home(c)
}

func (r *Router) ReleaseEnv(c echo.Context) error {
	user := c.Get("username").(string)
	env := c.Param("env")
	err := r.s.ReleaseEnvironment(env, user)
	if err != nil {
		return ReturnError(c, 400, err)
	}
	return r.Home(c)
}

func (r *Router) Login(c echo.Context) error {
	type loginParams struct {
		Name     string `json:"name" form:"name"`
		Password string `json:"password" form:"password"`
	}
	params := new(loginParams)
	if err := c.Bind(params); err != nil {
		return ReturnError(c, 400, err)
	}

	token, err := r.a.Authenticate(params.Name, params.Password)
	if err != nil {
		return ReturnError(c, 400, err)
	}

	c.SetCookie(&http.Cookie{Name: "token", Value: token})

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func (r *Router) Logout(c echo.Context) error {
	cookie, err := c.Cookie("token")
	if err != nil || cookie == nil {
		return ReturnError(c, 400, err)
	}
	cookie.Name = "token"
	cookie.Expires = time.Now()
	c.SetCookie(cookie)
	return c.Redirect(http.StatusMovedPermanently, "/")
}

func (r *Router) ResetPasswordForm(c echo.Context) error {
	return components.ResetPassword().Render(c.Request().Context(), c.Response().Writer)
}

func (r *Router) ResetPassword(c echo.Context) error {
	user := c.Get("username").(string)

	type resetPasswordParams struct {
		OldPassword     string `json:"oldPassword,omitempty" form:"oldPassword"`
		Password        string `json:"password,omitempty" form:"password"`
		PasswordConfirm string `json:"passwordConfirm,omitempty" form:"passwordConfirm"`
	}
	params := new(resetPasswordParams)
	if err := c.Bind(params); err != nil {
		return ReturnError(c, 400, err)
	}

	if params.Password != params.PasswordConfirm {
		return ReturnError(c, 400, errors.New("password confirmation error"))
	}

	err := r.a.ResetPassword(user, params.OldPassword, params.Password)
	if err != nil {
		return ReturnError(c, 400, err)
	}
	return r.Logout(c)
}

func Login(c echo.Context) error {
	return components.Login().Render(c.Request().Context(), c.Response().Writer)
}

func ReturnError(c echo.Context, status int, err error) error {
	c.Response().Header().Add("HX-Retarget", "#errors")
	c.Response().Header().Add("HX-Reswap", "innerHTML")
	return components.Error(status, err).Render(c.Request().Context(), c.Response().Writer)
}
