package main

import (
	"fmt"
	"github.com/bakyazi/envmutex"
	"github.com/labstack/echo/v4"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e := echo.New()
	envmutex.Init(e)
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", os.Getenv("LISTEN_ADDR"), port)))
}
