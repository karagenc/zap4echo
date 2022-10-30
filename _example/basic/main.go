package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/tomruk/zap4echo"
	"go.uber.org/zap"
)

/*
Try doing:
	curl http://127.0.0.1:8000
and
	curl http://127.0.0.1:8000/panic
*/

const addr = "127.0.0.1:8000"

func main() {
	e := echo.New()
	log, _ := zap.NewDevelopment()

	e.Use(
		zap4echo.Logger(log),
		zap4echo.Recover(log),
	)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Greetings!\n")
	})

	e.GET("/panic", func(c echo.Context) error {
		// This if is to prevent `unreachable` warning of linter
		if true {
			panic("intentional.")
		}
		return nil
	})

	err := e.Start(addr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
