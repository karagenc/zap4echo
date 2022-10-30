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
	assert.Equal(t, oops, l.ContextMap()["error"].(string))
	assert.Equal(t, "GET", l.ContextMap()["method"].(string))
	assert.Equal(t, "/panic", l.ContextMap()["path"].(string))
	assert.Equal(t, "192.0.2.1", l.ContextMap()["client_ip"].(string))

	assert.Nil(t, l.ContextMap()["request_id"])
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

func TestRecoverWithHandleError(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	const oops = "Oops, I did it again, I played with your heart"
	oopsErr := fmt.Errorf(oops)

	handleError := func(c echo.Context, err error) {
		assert.Equal(t, oopsErr, err)
		wg.Done()
	}

	config := RecoverConfig{
		HandleError: handleError,
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
