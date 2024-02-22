package main

import (
	"github.com/bakyazi/envmutex"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	envmutex.Init(e)
	e.Logger.Fatal(e.Start(":8080"))
}
