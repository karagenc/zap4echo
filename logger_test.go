package zap4echo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestDefaultLogger(t *testing.T) {
	log, logs := createTestZapLogger()
	m := Logger(log)
	e := createTestEcho(m)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(echo.HeaderXRequestID, "1337")
			return next(c)
		}
	})

	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello!")
	})

	r := httptest.NewRequest("GET", "/hello", nil)
	r.Host = "192.168.10.60:5252"
	r.Header.Set("User-Agent", "AnHTTPClient")
	r.Header.Set("Referer", "http://192.0.2.10")

	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	l := logs.All()[0]
	assert.Equal(t, zapcore.InfoLevel, l.Level)
	assert.Equal(t, DefaultLoggerMsg, l.Message)
	assert.Equal(t, "HTTP/1.1", l.ContextMap()["proto"].(string))
	assert.Equal(t, "192.168.10.60:5252", l.ContextMap()["host"].(string))
	assert.Equal(t, "GET", l.ContextMap()["method"].(string))
	assert.Equal(t, int64(http.StatusOK), l.ContextMap()["status"].(int64))
	assert.Equal(t, int64(6), l.ContextMap()["response_size"].(int64))

	_, ok := l.ContextMap()["latency"].(time.Duration)
	assert.Equal(t, true, ok)

	assert.Equal(t, http.StatusText(http.StatusOK), l.ContextMap()["status_text"].(string))
	assert.Equal(t, "192.0.2.1", l.ContextMap()["client_ip"].(string))
	assert.Equal(t, "AnHTTPClient", l.ContextMap()["user_agent"].(string))
	assert.Equal(t, "/hello", l.ContextMap()["path"].(string))
	assert.Equal(t, "1337", l.ContextMap()["request_id"].(string))
	assert.Equal(t, "http://192.0.2.10", l.ContextMap()["referer"].(string))

	// Caller is omitted by default unless 'IncludeCaller' is set to true.
	assert.Equal(t, false, l.Caller.Defined)
}

func TestDefaultLoggerWithDifferentLevels(t *testing.T) {
	log, logs := createTestZapLogger()
	m := Logger(log)
	e := createTestEcho(m)

	e.GET("/serviceunavailable", func(c echo.Context) error {
		return c.NoContent(http.StatusServiceUnavailable)
	})

	e.GET("/teapot", func(c echo.Context) error {
		return c.NoContent(http.StatusTeapot)
	})

	e.GET("/seeother", func(c echo.Context) error {
		return c.NoContent(http.StatusSeeOther)
	})

	e.GET("/earlyhints", func(c echo.Context) error {
		return c.NoContent(http.StatusEarlyHints)
	})

	r := httptest.NewRequest("GET", "/serviceunavailable", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)

	l := logs.All()[0]
	assert.Equal(t, zapcore.ErrorLevel, l.Level)

	r = httptest.NewRequest("GET", "/teapot", nil)
	w = httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res = w.Result()
	assert.Equal(t, http.StatusTeapot, res.StatusCode)

	l = logs.All()[1]
	assert.Equal(t, zapcore.WarnLevel, l.Level)

	r = httptest.NewRequest("GET", "/seeother", nil)
	w = httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res = w.Result()
	assert.Equal(t, http.StatusSeeOther, res.StatusCode)

	l = logs.All()[2]
	assert.Equal(t, zapcore.InfoLevel, l.Level)

	r = httptest.NewRequest("GET", "/earlyhints", nil)
	w = httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res = w.Result()
	assert.Equal(t, http.StatusEarlyHints, res.StatusCode)

	l = logs.All()[3]
	assert.Equal(t, zapcore.InfoLevel, l.Level)
}

func TestDefaultLoggerWithErrorOnly(t *testing.T) {
	config := LoggerConfig{
		ErrorOnly: true,
	}

	log, logs := createTestZapLogger()
	m := LoggerWithConfig(log, config)
	e := createTestEcho(m)

	e.GET("/success", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	e.GET("/500", func(c echo.Context) error {
		return c.NoContent(http.StatusInternalServerError)
	})

	e.GET("/error", func(c echo.Context) error {
		return fmt.Errorf("intentional")
	})

	r := httptest.NewRequest("GET", "/success", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	assert.Equal(t, 0, logs.Len())

	r = httptest.NewRequest("GET", "/error", nil)
	w = httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res = w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	assert.Equal(t, 1, logs.Len())

	r = httptest.NewRequest("GET", "/500", nil)
	w = httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res = w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	assert.Equal(t, 2, logs.Len())
}

func TestDefaultLoggerWithSkipper(t *testing.T) {
	config := LoggerConfig{
		Skipper: func(c echo.Context) bool {
			if c.Param("skip") == "true" {
				return true
			}
			return false
		},
	}

	log, logs := createTestZapLogger()
	m := LoggerWithConfig(log, config)
	e := createTestEcho(m)

	e.GET("/:skip", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/true", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	assert.Equal(t, 0, logs.Len())

	r = httptest.NewRequest("GET", "/false", nil)
	w = httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res = w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	assert.Equal(t, 1, logs.Len())
}

func TestDefaultLoggerWithCustomMsg(t *testing.T) {
	const customMsg = "This is the custom message."
	config := LoggerConfig{
		CustomMsg: customMsg,
	}

	log, logs := createTestZapLogger()
	m := LoggerWithConfig(log, config)
	e := createTestEcho(m)

	e.GET("/", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	assert.Equal(t, customMsg, logs.All()[0].Message)
}

func TestLoggerWithCustomMsg(t *testing.T) {
	const customMsg = "31337"
	config := LoggerConfig{
		CustomMsg: customMsg,
	}

	log, logs := createTestZapLogger()
	m := LoggerWithConfig(log, config)
	e := createTestEcho(m)

	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello!")
	})

	r := httptest.NewRequest("GET", "/hello", nil)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	l := logs.All()[0]
	assert.Equal(t, customMsg, l.Message)
}

func TestDefaultLoggerWithIncludeCaller(t *testing.T) {
	config := LoggerConfig{
		IncludeCaller: true,
	}

	log, logs := createTestZapLogger()
	m := LoggerWithConfig(log, config)
	e := createTestEcho(m)

	e.GET("/", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	assert.Equal(t, true, logs.All()[0].Caller.Defined)
}

func TestLoggerWithEverythingOmitted(t *testing.T) {
	config := LoggerConfig{
		OmitStatusText: true,
		OmitClientIP:   true,
		OmitUserAgent:  true,
		OmitPath:       true,
		OmitRequestID:  true,
		OmitReferer:    true,
	}

	log, logs := createTestZapLogger()
	m := LoggerWithConfig(log, config)
	e := createTestEcho(m)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(echo.HeaderXRequestID, "1337")
			return next(c)
		}
	})

	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello!")
	})

	r := httptest.NewRequest("GET", "/hello", nil)
	r.Host = "192.168.10.60:5252"
	r.Header.Set("User-Agent", "AnHTTPClient")
	r.Header.Set("Referer", "http://192.0.2.10")

	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	l := logs.All()[0]
	assert.Equal(t, zapcore.InfoLevel, l.Level)
	assert.Equal(t, DefaultLoggerMsg, l.Message)
	assert.Equal(t, "HTTP/1.1", l.ContextMap()["proto"].(string))
	assert.Equal(t, "192.168.10.60:5252", l.ContextMap()["host"].(string))
	assert.Equal(t, "GET", l.ContextMap()["method"].(string))
	assert.Equal(t, int64(http.StatusOK), l.ContextMap()["status"].(int64))
	assert.Equal(t, int64(6), l.ContextMap()["response_size"].(int64))

	_, ok := l.ContextMap()["latency"].(time.Duration)
	assert.Equal(t, true, ok)

	assert.Nil(t, l.ContextMap()["status_text"])
	assert.Nil(t, l.ContextMap()["client_ip"])
	assert.Nil(t, l.ContextMap()["user_agent"])
	assert.Nil(t, l.ContextMap()["path"])
	assert.Nil(t, l.ContextMap()["request_id"])
	assert.Nil(t, l.ContextMap()["referer"])
}

func TestLoggerWithCustomRequestIDHeader(t *testing.T) {
	const requestID = "31337"
	const requestIDHeader = "My1337RequestID"
	config := LoggerConfig{
		CustomRequestIDHeader: requestIDHeader,
	}

	log, logs := createTestZapLogger()
	m := LoggerWithConfig(log, config)
	e := createTestEcho(m)

	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello!")
	})

	r := httptest.NewRequest("GET", "/hello", nil)
	r.Header.Set(requestIDHeader, requestID)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	l := logs.All()[0]
	assert.Equal(t, requestID, l.ContextMap()["request_id"].(string))
}

// This is almost the same as `TestLoggerWithCustomRequestIDHeader`,
// but the difference is that request ID is set by server instead of client.
func TestLoggerWithCustomRequestIDHeader2(t *testing.T) {
	const requestID = "31337"
	const requestIDHeader = "My1337RequestID"
	config := LoggerConfig{
		CustomRequestIDHeader: requestIDHeader,
	}

	log, logs := createTestZapLogger()
	m := LoggerWithConfig(log, config)
	e := createTestEcho(m)

	e.GET("/hello", func(c echo.Context) error {
		c.Response().Header().Set(requestIDHeader, requestID)
		return c.String(http.StatusOK, "Hello!")
	})

	r := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	l := logs.All()[0]
	assert.Equal(t, requestID, l.ContextMap()["request_id"].(string))
}

func TestDefaultLoggerWithFieldAdder(t *testing.T) {
	config := LoggerConfig{
		FieldAdder: func(c echo.Context) []zapcore.Field {
			return []zapcore.Field{
				zap.String("hello", "world!"),
				zap.Bool("b", true),
			}
		},
	}

	log, logs := createTestZapLogger()
	m := LoggerWithConfig(log, config)
	e := createTestEcho(m)

	e.GET("/", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	l := logs.All()[0]
	assert.Equal(t, "world!", l.ContextMap()["hello"].(string))
	assert.Equal(t, true, l.ContextMap()["b"].(bool))
}

func createTestEcho(middleware echo.MiddlewareFunc) *echo.Echo {
	e := echo.New()
	e.Debug = true
	e.Use(middleware)
	return e
}

func createTestZapLogger() (log *zap.Logger, logs *observer.ObservedLogs) {
	observed, logs := observer.New(zap.InfoLevel)
	log = zap.New(observed, zap.AddCaller())
	return log, logs
}
