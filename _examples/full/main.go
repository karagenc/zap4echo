package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/karagenc/zap4echo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/*
Try doing:
	curl http://127.0.0.1:8000
	curl http://127.0.0.1:8000/greet/John
	curl http://127.0.0.1:8000/nolog
	curl http://127.0.0.1:8000/panic
	curl -H 'My-Request-Id: 31337' http://127.0.0.1:8000
*/

const addr = "127.0.0.1:8000"

func main() {
	e := echo.New()
	log, _ := zap.NewDevelopment()

	loggerConfig := zap4echo.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			if c.Request().URL.Path == "/nolog" {
				return true
			}
			return false
		},

		CustomMsg:             "Request received!",
		IncludeCaller:         true,
		CustomRequestIDHeader: "My-Request-Id",

		FieldAdder: func(c echo.Context) []zapcore.Field {
			name := c.Param("name")
			if name != "" {
				return []zapcore.Field{
					zap.String("name", name),
				}
			}
			return nil
		},
	}

	recoverConfig := zap4echo.RecoverConfig{
		CustomMsg:             "Panic happened!",
		StackTrace:            true,
		StackTraceSize:        4 << 10, // 4 kb
		CustomRequestIDHeader: "My-Request-Id",

		ErrorHandler: func(c echo.Context, err error) {
			fmt.Printf("Panic :( Error: %v\n", err)
		},
	}

	e.Use(
		zap4echo.LoggerWithConfig(log, loggerConfig),
		zap4echo.RecoverWithConfig(log, recoverConfig),
	)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello!\n")
	})

	e.GET("/greet/:name", func(c echo.Context) error {
		name := c.Param("name")
		return c.String(http.StatusOK, fmt.Sprintf("Greetings %s.\n", name))
	})

	e.GET("/nolog", func(c echo.Context) error {
		return c.String(http.StatusOK, "This will not be logged.\n")
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
