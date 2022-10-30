package zap4echo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestDefaultRecover(t *testing.T) {
	log, logs := createTestZapLogger()
	m := Recover(log)
	e := createTestEcho(m)

	const oops = "Oops, I did it again, I played with your heart"

	e.GET("/panic", func(c echo.Context) error {
		if true {
			panic(oops)
		}
		return nil
	})

	r := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	l := logs.All()[0]
	assert.Equal(t, zapcore.ErrorLevel, l.Level)
	assert.Equal(t, DefaultRecoverMsg, l.Message)
	assert.Equal(t, oops, l.ContextMap()["error"].(string))
	assert.Equal(t, "GET", l.ContextMap()["method"].(string))
	assert.Equal(t, "/panic", l.ContextMap()["path"].(string))
	assert.Equal(t, "192.0.2.1", l.ContextMap()["client_ip"].(string))

	assert.Nil(t, l.ContextMap()["request_id"])
}

func TestRecoverWithCustomMsg(t *testing.T) {
	const customMsg = "31337"
	config := RecoverConfig{
		CustomMsg: customMsg,
	}

	log, logs := createTestZapLogger()
	m := RecoverWithConfig(log, config)
	e := createTestEcho(m)

	const oops = "Oops, I did it again, I played with your heart"

	e.GET("/panic", func(c echo.Context) error {
		if true {
			panic(oops)
		}
		return nil
	})

	r := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	l := logs.All()[0]

	assert.Equal(t, customMsg, l.Message)
}

func TestRecoverWithStackTrace(t *testing.T) {
	config := RecoverConfig{
		StackTrace: true,
	}

	log, logs := createTestZapLogger()
	// Disable stacktrace
	log = log.WithOptions(zap.AddStacktrace(zap.FatalLevel + 1))

	m := RecoverWithConfig(log, config)
	e := createTestEcho(m)

	const oops = "Oops, I did it again, I played with your heart"

	e.GET("/panic", func(c echo.Context) error {
		if true {
			panic(oops)
		}
		return nil
	})

	r := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	l := logs.All()[0]
	stacktrace := l.ContextMap()["stacktrace"].(string)

	assert.Equal(t, oops, l.ContextMap()["error"].(string))
	assert.Contains(t, stacktrace, "zap4echo")
}

func TestRecoverWithStackTraceSize(t *testing.T) {
	config := RecoverConfig{
		StackTrace:     true,
		StackTraceSize: 10,
	}

	log, logs := createTestZapLogger()
	m := RecoverWithConfig(log, config)
	e := createTestEcho(m)

	const oops = "Oops, I did it again, I played with your heart"

	e.GET("/panic", func(c echo.Context) error {
		if true {
			panic(oops)
		}
		return nil
	})

	r := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	l := logs.All()[0]
	stacktrace := l.ContextMap()["stacktrace"].(string)

	assert.Equal(t, oops, l.ContextMap()["error"].(string))
	assert.Equal(t, 10, len(stacktrace))
}

func TestRecoverWithCustomRequestIDHeader(t *testing.T) {
	const requestID = "31337"
	const requestIDHeader = "My1337RequestID"
	config := RecoverConfig{
		CustomRequestIDHeader: requestIDHeader,
	}

	log, logs := createTestZapLogger()
	m := RecoverWithConfig(log, config)
	e := createTestEcho(m)

	const oops = "Oops, I did it again, I played with your heart"

	e.GET("/panic", func(c echo.Context) error {
		if true {
			panic(oops)
		}
		return nil
	})

	r := httptest.NewRequest("GET", "/panic", nil)
	r.Header.Set(requestIDHeader, requestID)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	l := logs.All()[0]

	assert.Equal(t, oops, l.ContextMap()["error"].(string))
	assert.Equal(t, requestID, l.ContextMap()["request_id"].(string))
}

func TestRecoverWithCustomRequestIDHeader2(t *testing.T) {
	const requestID = "31337"
	const requestIDHeader = "My1337RequestID"
	config := RecoverConfig{
		CustomRequestIDHeader: requestIDHeader,
	}

	log, logs := createTestZapLogger()
	m := RecoverWithConfig(log, config)
	e := createTestEcho(m)

	const oops = "Oops, I did it again, I played with your heart"

	e.GET("/panic", func(c echo.Context) error {
		c.Response().Header().Set(requestIDHeader, requestID)
		if true {
			panic(oops)
		}
		return nil
	})

	r := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	l := logs.All()[0]

	assert.Equal(t, oops, l.ContextMap()["error"].(string))
	assert.Equal(t, requestID, l.ContextMap()["request_id"].(string))
}

func TestRecoverWithFieldAdder(t *testing.T) {
	config := RecoverConfig{
		FieldAdder: func(c echo.Context, err error) []zap.Field {
			return []zapcore.Field{
				zap.String("hello", "world!"),
				zap.Bool("b", true),
			}
		},
	}

	log, logs := createTestZapLogger()
	m := RecoverWithConfig(log, config)
	e := createTestEcho(m)

	const oops = "Oops, I did it again, I played with your heart"

	e.GET("/panic", func(c echo.Context) error {
		if true {
			panic(oops)
		}
		return nil
	})

	r := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	l := logs.All()[0]
	assert.Equal(t, oops, l.ContextMap()["error"].(string))
	assert.Equal(t, "world!", l.ContextMap()["hello"].(string))
	assert.Equal(t, true, l.ContextMap()["b"].(bool))
}

func TestRecoverWithErrorHandler(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	const oops = "Oops, I did it again, I played with your heart"
	oopsErr := fmt.Errorf(oops)

	handleError := func(c echo.Context, err error) {
		assert.Equal(t, oopsErr, err)
		wg.Done()
	}

	config := RecoverConfig{
		ErrorHandler: handleError,
	}

	log, logs := createTestZapLogger()
	m := RecoverWithConfig(log, config)
	e := createTestEcho(m)

	e.GET("/panic", func(c echo.Context) error {
		if true {
			panic(oopsErr)
		}
		return nil
	})

	r := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	l := logs.All()[0]

	assert.Equal(t, oops, l.ContextMap()["error"].(string))
	wg.Wait()
}
